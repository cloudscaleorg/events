package events

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/ldelossa/goframework/backoff"
	"github.com/ldelossa/goframework/chkctx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	etcd "go.etcd.io/etcd/clientv3"
)

// ReduceFunc is called on each event ingress. provided by the caller on construction
// snapshot will be true iff event is the first event reduced in the snapshot state.
type ReduceFunc func(e *etcd.Event, snapshot bool)

// State is the enum identifying the possible states of a Listener
type State int

func (s *State) ToString() string {
	m := map[State]string{
		0: "Terminal",
		1: "Buffer",
		2: "Snapshot",
		3: "Listening",
	}
	return m[*s]
}

// States and explanations
const (
	// used to express the listener is exiting
	Terminal State = iota
	// opens a watch channel with etcd at the configured prefix. buffers events for replay
	Buffer
	// queries the configured prefix for current k/vs.
	Snapshot
	// calls ReduceFunc for any existing events on the watch channel and all subsequent.
	Listening
)

type stateFunc func(ctx context.Context, l *Listener) State

var stateToStateFunc = map[State]stateFunc{
	Buffer:    buffer,
	Snapshot:  snapshot,
	Listening: listening,
}

// Listener maintains an active connection to etcd via a watch channel. for each event
// that is ingressed the provided ReduceFunc will be called.
type Listener struct {
	Fencer
	Prefix  string
	F       ReduceFunc
	Client  *etcd.Client
	Backoff backoff.BackOff
	// holds our State type
	state atomic.Value
	// channel we listen to events on
	eC <-chan *etcd.Event
	// a logger with per listener context
	logger zerolog.Logger
}

// NewListener is a constructor for an event.Listener
func NewListener(opts *Opts) (*Listener, error) {
	if err := opts.Parse(); err != nil {
		return nil, err
	}

	l := &Listener{
		Fencer:  NewFencer(),
		Prefix:  opts.Prefix,
		F:       opts.F,
		Client:  opts.Client,
		Backoff: backoff.NewExpoBackoff(opts.MaxBackoff),
		logger:  log.With().Str("component", "listener").Str("prefix", opts.Prefix).Logger(),
	}
	l.setState(Buffer)
	return l, nil
}

// Listen kicks off the Listener in it's own go routine. this method is
// non-blocking. cancel the ctx to stop the Listener.
func (l *Listener) Listen(ctx context.Context) {
	l.logger.Debug().Msg("listener started")
	go l.run(ctx)
}

// run transitions through the listener's states
func (l *Listener) run(ctx context.Context) {
	var state State
	for {
		state = stateToStateFunc[l.getState()](ctx, l)
		l.Backoff.Do()
		if ok, _ := chkctx.Check(ctx); ok {
			l.setState(Terminal)
			return
		}
		l.setState(state)
	}
}

func (l *Listener) getState() State {
	s := l.state.Load()
	return s.(State)
}

func (l *Listener) setState(s State) {
	l.state.Store(s)
	l.logger.Info().Msgf("state change: %v", s.ToString())
}

// Ready will block until the Listener is in Listening state, Terminal state,
// or the provided ctx is canceled.
//
// If provided ctx is canceled or Terminal state
// encountered an error is returned.
func (l *Listener) Ready(ctx context.Context) error {
	if l.getState() == Listening {
		return nil
	}

	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return ErrReadyTimeout
		case <-t.C:
			if l.getState() == Terminal {
				return ErrListenerStopped
			}
			if l.getState() == Listening {
				return nil
			}
		}
	}
}

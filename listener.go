package events

import (
	"context"
	"fmt"
	"sync"

	"github.com/ldelossa/goframework/backoff"
	"github.com/ldelossa/goframework/chkctx"
	etcd "go.etcd.io/etcd/clientV3"
)

// ReduceFunc is called on each event ingress. provided by the caller on construction
type ReduceFunc func(e *etcd.Event)

// State is the enum identifying the possible states of a Listener
type State int

const (
	// used to express the listener is existing
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
	stateMu *sync.RWMutex
	state   State
	ready   *sync.Cond
	// the configured prefix to listen on
	eC <-chan *etcd.Event
	// stops the Listen() method from spawning multiple go routines
	active bool
}

// NewListener is a constructor for an event.Listener
func NewListener(opts Opts) (*Listener, error) {
	if err := opts.Parse(); err != nil {
		return nil, err
	}

	backoff := backoff.NewExpoBackoff(opts.MaxBackoff)
	stateMu := &sync.RWMutex{}
	return &Listener{
		Fencer:  NewFencer(),
		Prefix:  opts.Prefix,
		F:       opts.F,
		Client:  opts.Client,
		Backoff: backoff,
		stateMu: stateMu,
		state:   Buffer,
		ready:   sync.NewCond(stateMu.RLocker()),
	}, nil
}

// Listen kicks off the Listener in it's own go routine. this method is
// non-blocking. cancel the ctx to stop the Listener.
func (l *Listener) Listen(ctx context.Context) {
	go l.run(ctx)
}

// run transitions through the listener's states
func (l *Listener) run(ctx context.Context) {
	var state State
	for {
		state = stateToStateFunc[l.state](ctx, l)
		l.Backoff.Do()

		if ok, _ := chkctx.Check(ctx); ok {
			l.setState(Terminal)
			// broadcast to unblock any callers blocked on Ready
			l.ready.Broadcast()
			return
		}

		l.setState(state)
	}
}

func (l *Listener) GetState() State {
	l.stateMu.RLock()
	state := l.state
	l.stateMu.RUnlock()
	return state
}

func (l *Listener) setState(s State) {
	l.stateMu.Lock()
	l.state = s
	l.stateMu.Unlock()

	l.ready.Broadcast()
}

// Ready will block until the Listener is in Listening state, Terminal state,
// or the provided ctx is canceled.
//
// If provided ctx is canceled or Terminal state
// encountered an error is returned.
func (l *Listener) Ready(ctx context.Context) error {
	l.ready.L.Lock()
	defer l.ready.L.Unlock()
	for {
		if done, err := chkctx.Check(ctx); done {
			return err
		}

		if l.state == Terminal {
			return fmt.Errorf("listener is returning. can no longer block on ready")
		}

		if l.state == Listening {
			return nil
		}
		l.ready.Wait()
	}
}

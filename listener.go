package events

import (
	"context"
	"sync"
	"time"

	"github.com/ldelossa/goframework/backoff"
	"github.com/ldelossa/goframework/chkctx"
	etcd "go.etcd.io/etcd/clientV3"
)

const (
	// how long lease manager will wait for etcd operations to complete
	defaultOPTimeout = 30 * time.Second
	// upper bound leaseManager will wait to retry when unsuccessfully obtaining a lease
	defaultMaxBackoff = 64 * time.Second
)

// ReduceFunc is called on each event ingress. provided by the caller on construction
type ReduceFunc func(e *etcd.Event)

// State is the enum identifying the possible states of a Listener
type State int

const (
	// stops the event listener
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

// Opts is configuration for a listener
type Opts struct {
	// the prefix to listen on
	Prefix string
	Client *etcd.Client
	F      ReduceFunc
	// max duration we will wait to retry etcd operations
	MaxBackoff time.Duration
	// max time we will wait for etcd operations to complete
	OPTimeout time.Duration
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
func NewListener(opts Opts) *Listener {
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
	}
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
			return
		}

		l.SetState(state)
	}
}

func (l *Listener) GetState() State {
	l.stateMu.RLock()
	state := l.state
	l.stateMu.RUnlock()
	return state
}

func (l *Listener) SetState(s State) {
	l.stateMu.Lock()
	l.state = s
	l.stateMu.Unlock()
}

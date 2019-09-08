package events

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	etcd "go.etcd.io/etcd/clientv3"
)

// Test_Listener_Ready_Listening confirms a caller
// of Ready() will unblock when the listener enters Listening state
func Test_Listener_Ready_Listening(t *testing.T) {
	opts := &Opts{
		Prefix: "null",
		Client: &etcd.Client{},
		F:      func(e *etcd.Event, snapshot bool) {},
	}

	l, err := NewListener(opts)
	assert.NoError(t, err)

	readyCTX, cancel := context.WithCancel(context.Background())
	doneChan := make(chan error, 0)
	go func(ctx context.Context, c chan error) {
		err := l.Ready(ctx)
		doneChan <- err
	}(readyCTX, doneChan)

	l.setState(Listening)

	testTOCTX, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	select {
	case <-testTOCTX.Done():
		t.Fatalf("test timed out. ready never unblocked")
	case e := <-doneChan:
		assert.NoError(t, e)
	}

}

// Test_Listener_Ready_Timeout confirms a caller
// of Ready() will unblock when the provided ctx times out
func Test_Listener_Ready_Timeout(t *testing.T) {

	opts := &Opts{
		Prefix: "null",
		Client: &etcd.Client{},
		F:      func(e *etcd.Event, snapshot bool) {},
	}

	l, err := NewListener(opts)
	assert.NoError(t, err)

	readyCTX, cancel := context.WithCancel(context.Background())
	doneChan := make(chan error, 0)
	go func(ctx context.Context, c chan error) {
		err := l.Ready(ctx)
		doneChan <- err
	}(readyCTX, doneChan)
	cancel()

	testTOCTX, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	select {
	case <-testTOCTX.Done():
		t.Fatalf("test timed out. ready never unblocked")
	case e := <-doneChan:
		assert.Error(t, e)
		assert.IsType(t, ErrReadyTimeout, e)
	}
}

// Test_Listener_Ready_Terminal confirms a caller of Ready() will
// unblock when the listener enters Terminal state
func Test_Listener_Ready_Terminal(t *testing.T) {
	opts := &Opts{
		Prefix: "null",
		Client: &etcd.Client{},
		F:      func(e *etcd.Event, snapshot bool) {},
	}

	l, err := NewListener(opts)
	assert.NoError(t, err)

	readyCTX, cancel := context.WithCancel(context.Background())
	doneChan := make(chan error, 0)
	go func(ctx context.Context, c chan error) {
		err := l.Ready(ctx)
		doneChan <- err
	}(readyCTX, doneChan)

	l.setState(Terminal)

	testTOCTX, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	select {
	case <-testTOCTX.Done():
		t.Fatalf("test timed out. ready never unblocked")
	case e := <-doneChan:
		assert.Error(t, e)
		assert.IsType(t, ErrListenerStopped, e)
	}

}

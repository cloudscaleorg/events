package events

import (
	"context"

	etcd "github.com/coreos/etcd/clientV3"
	"github.com/coreos/etcd/mvcc/mvccpb"
)

// snapshot does a GET of all the KV's at the configured prefix
// and popoulates the fence map with initial values
func snapshot(ctx context.Context, l *Listener) State {
	getCTX, cancel := context.WithTimeout(ctx, defaultOPTimeout)
	defer cancel()

	resp, err := l.Client.KV.Get(getCTX, l.Prefix, etcd.WithPrefix())
	if err != nil {
		// restart state machine
		l.Backoff.Increment()
		return Buffer
	}

	for _, kv := range resp.Kvs {
		if !l.Fencer.Fence(kv) {
			// synthesize an event. a get will not return deleted KVs
			l.F(&etcd.Event{
				Type: mvccpb.PUT,
				Kv:   kv,
			})
		}
	}

	return Listening
}

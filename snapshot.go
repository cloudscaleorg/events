package events

import (
	"context"

	etcd "go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

// snapshot does a GET of all the KV's at the configured prefix
// and popoulates the fence map with initial values
func snapshot(ctx context.Context, l *Listener) State {
	getCTX, cancel := context.WithTimeout(ctx, defaultOPTimeout)
	defer cancel()

	resp, err := l.Client.KV.Get(getCTX, l.Prefix, etcd.WithPrefix())
	if err != nil {
		l.logger.Error().Msgf("error getting etcd keys at given prefix. backing off...: %v", err)
		l.Backoff.Increment()
		// restart state machine after backoff
		return Buffer
	}

	var snapshot = true
	for _, kv := range resp.Kvs {
		if !l.Fence(kv) {
			// synthesize an event. a get will not return deleted KVs
			l.F(&etcd.Event{
				Type: mvccpb.PUT,
				Kv:   kv,
			}, snapshot)
			snapshot = false
		}
	}

	l.Backoff.Reset()
	l.logger.Debug().Msg("snapshot successfully created")
	return Listening
}

package events

import (
	"context"

	"github.com/ldelossa/goframework/chkctx"
	"github.com/rs/zerolog/log"
	etcd "go.etcd.io/etcd/clientv3"
)

// Stream demultiplexes events from an etcd.WatchResponse and delivers them to the returned channel.
// provides channel semantics, such as ranging, to the caller.
// remember to cancel context to not leak go routines
func Stream(ctx context.Context, wC etcd.WatchChan) <-chan *etcd.Event {
	eC := make(chan *etcd.Event, 1024)

	go func(ctx context.Context, ec chan *etcd.Event) {
		// this unblocks any callers ranging on ec
		defer close(ec)

		// etcd client will close this channel if error occurs
		for wResp := range wC {
			if ok, err := chkctx.Check(ctx); ok {
				log.Info().Str("component", "Stream").Msgf("stream ctx canceled. returning: %v", err)
				return
			}

			if wResp.Canceled {
				log.Info().Str("component", "Stream").Msgf("watch channel error encountered. returning: %v", wResp.Err())
				return
			}

			for _, event := range wResp.Events {
				eC <- event
			}
		}
	}(ctx, eC)

	return eC
}

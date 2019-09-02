package events

import (
	"context"

	"github.com/ldelossa/goframework/chkctx"
	"github.com/rs/zerolog/log"
	etcd "go.etcd.io/etcd/clientV3"
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
		for watchResp := range wC {
			if ok, err := chkctx.Check(ctx); ok {
				log.Info().Str("component", "Streamer").Msgf("streamer ctx canceled. returning: %v", err)
				return
			}

			if watchResp.Canceled {
				log.Info().Str("component", "Streamer").Msgf("watch channel error encountered. returning: %v")
			}

			for _, event := range watchResp.Events {
				eC <- event
			}
		}
	}(ctx, eC)

	return eC
}

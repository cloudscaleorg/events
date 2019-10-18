package events

import (
	"context"

	etcd "go.etcd.io/etcd/clientv3"
)

func buffer(ctx context.Context, l *Listener) State {
	wC := l.Client.Watcher.Watch(ctx, l.Prefix, etcd.WithPrefix(), etcd.WithPrevKV())
	l.eC = Stream(ctx, wC)
	l.logger.Debug().Msg("now buffering events")
	return Snapshot
}

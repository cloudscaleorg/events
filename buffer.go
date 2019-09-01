package events

import "context"

func buffer(ctx context.Context, l *Listener) State {
	wC := l.Client.Watcher.Watch(ctx, l.Prefix)
	l.eC = Stream(ctx, wC)
	return Snapshot
}

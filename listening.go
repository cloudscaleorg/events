package events

import "context"

// listening reduces any messages previously in the event channel
// and reduces all subsequent ones
func listening(ctx context.Context, l *Listener) State {
	for event := range l.eC {
		if !l.Fence(event.Kv) {
			l.F(event)
		}
	}

	return Buffer
}

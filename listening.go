package events

import "context"

// listening reduces any messages previously in the event channel
// and reduces all subsequent ones
func listening(ctx context.Context, l *Listener) State {
	// we do not listen on the ctx directly in this state
	// if the ctx is canceled l.eC channel will be closed
	// and loop with unblock. see stream.go
	for event := range l.eC {
		if !l.Fence(event.Kv) {
			l.F(event)
		}
	}

	return Buffer
}

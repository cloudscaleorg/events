package events

import "errors"

var (
	ErrReadyTimeout    = errors.New("ctx deadline or canceled by caller")
	ErrListenerStopped = errors.New("event listener is in a terminal state")
)

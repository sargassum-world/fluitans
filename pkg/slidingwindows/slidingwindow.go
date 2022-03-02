package slidingwindows

import (
	"time"
)

// Window represents a fixed-window.
type Window interface {
	// Start returns the start boundary.
	Start() time.Time

	// Count returns the accumulated count.
	Count() int64

	// AddCount increments the accumulated count by n.
	AddCount(n int64)

	// Reset sets the state of the window with the given settings.
	Reset(s time.Time, c int64)

	// Sync tries to exchange data between the window and the central
	// datastore at time now, to keep the window's count up-to-date.
	Sync(now time.Time)
}

// StopFunc stops the window's sync behavior.
type StopFunc func()

// NewWindow creates a new window, and returns a function to stop
// the possible sync behavior within it.
type NewWindow func() (Window, StopFunc)

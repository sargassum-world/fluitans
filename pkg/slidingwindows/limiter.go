package slidingwindows

import (
	"sync"
	"time"
)

type Limiter struct {
	capacity int64
	duration time.Duration

	mu sync.Mutex

	curr Window
	prev Window
}

// NewLimiter creates a new limiter, and returns a function to stop
// the possible sync behavior within the current window.
func NewLimiter(duration time.Duration, capacity int64, newWindow NewWindow) (*Limiter, StopFunc) {
	currWin, currStop := newWindow()

	// The previous window is static (i.e. no add changes will happen within it),
	// so we always create it as an instance of LocalWindow.
	//
	// In this way, the whole limiter, despite containing two windows, now only
	// consumes at most one goroutine for the possible sync behavior within
	// the current window.
	prevWin, _ := NewLocalWindow()

	lim := &Limiter{
		duration: duration,
		capacity: capacity,
		curr:     currWin,
		prev:     prevWin,
	}

	return lim, currStop
}

// Capacity returns the maximum number of events permitted to happen during one
// window duration. Note that the capacity is defined to be read-only; if you
// need to chance the capacity, create a new limiter with a new capacity instead.
func (lim *Limiter) Capacity() int64 {
	return lim.capacity
}

// Duration returns the time duration of one window duration. Note that the duration
// is defined to be read-only; if you need to change the duration, create a new
// limiter with a new duration instead.
func (lim *Limiter) Duration() time.Duration {
	return lim.duration
}

// advance updates the current/previous windows resulting from the passage of time.
func (lim *Limiter) advance(now time.Time) {
	// Calculate the start boundary of the expected current-window.
	newCurrStart := now.Truncate(lim.duration)

	diffDuration := newCurrStart.Sub(lim.curr.Start()) / lim.duration
	if diffDuration < 1 {
		return
	}

	// The current-window is at least one-window-duration behind the expected one.
	newPrevCount := int64(0)
	if diffDuration == 1 {
		// The new previous-window will overlap with the old current-window,
		// so it inherits the count.
		//
		// Note that the count here may be inaccurate when a synchronizer is needed,
		// since it is only a SNAPSHOT of the current-window's count, which in itself
		// tends to be inaccurate due to the asynchronous nature of the sync behavior.
		newPrevCount = lim.curr.Count()
	}
	lim.prev.Reset(newCurrStart.Add(-lim.duration), newPrevCount)

	// The new current-window always has zero count.
	lim.curr.Reset(newCurrStart, 0)
}

// estimateCount is a lock-free, mutation-free version of EstimateCountAt for
// internal use by other methods which already take responsibility for advancing
// the windows and managing locks.
func (lim *Limiter) estimateCount(now time.Time) int64 {
	lim.advance(now)
	elapsed := now.Sub(lim.curr.Start())
	weight := float64(lim.duration-elapsed) / float64(lim.duration)
	return int64(weight*float64(lim.prev.Count())) + lim.curr.Count()
}

// EstimateCount returns an approximate estimate at time now of the number of
// rate-limited events which have occurred over the sliding window. It makes no
// guarantees about how long the count will be valid for, or (when a synchronizer
// is used) how accurate the count is at any instant, so it should only be used
// in situations where inaccuracies are tolerable.
// Warning: it advances the internal windows, so it shouldn't be used to estimate
// counts at hypothetical times in the future!
func (lim *Limiter) EstimateCount(now time.Time) int64 {
	lim.mu.Lock()
	defer lim.mu.Unlock()

	return lim.estimateCount(now)
}

// EstimateFillRatio returns an approximate estimate at time now of the how close
// the sliding window is to its limit, as a ratio between 0 and 1. It makes no
// guarantees about how long the ratio will be valid for, or (when a synchronizer
// is used) how accurate the ratio is at any instant, so it should only be used
// in situations where inaccuracies are tolerable.
// Warning: it advances the internal windows, so it shouldn't be used to estimate
// counts at hypothetical times in the future!
func (lim *Limiter) EstimateFillRatio(now time.Time) float64 {
	return float64(lim.EstimateCount(now)) / float64(lim.capacity)
}

// EstimateWaitDuration returns a reasonable upper bound on the amount of time
// that must pass before the limiter at time now will have capacity to add n items,
// assuming no other events are added in the meantime.
func (lim *Limiter) EstimateWaitDuration(now time.Time, n int64) time.Duration {
	available := lim.capacity - lim.EstimateCount(now)
	excess := n - available
	if excess <= 0 {
		return 0
	}
	return time.Duration(float64(excess) * float64(lim.duration) / float64(lim.capacity))
}

// MaybeAllowed reports whether the limiter at time now could possibly stay
// at/below capacity (with the caveat that either it might reach capacity asynchronously
// before the next method call, or (when a synchronizer is used) it might already be
// at capacity) or if n events were to be added to it.
// It should only be used to short-circuit work when the limiter could not possibly
// stay within its capacity limits.
// Warning: it advances the internal windows, so it shouldn't be used to estimate
// allowances at hypothetical times in the future!
func (lim *Limiter) MaybeAllowed(now time.Time, n int64) bool {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	defer lim.curr.Sync(now)

	return lim.estimateCount(now)+n <= lim.capacity
}

// TryAdd reports whether n events are allowed to happen at time now, and if so,
// adds them to the sliding window in an all-or-nothing manner.
func (lim *Limiter) TryAdd(now time.Time, n int64) bool {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	defer lim.curr.Sync(now)

	allowed := lim.estimateCount(now)+n <= lim.capacity

	if allowed {
		lim.curr.AddCount(n)
	}
	return allowed
}

// AddUpTo immediately adds up to n events to the limiter at time now, stopping
// only when capacity is exhausted, and returns the number of events added.
// It differs in that TryAdd is all-or-nothing, while AddUpTo is a best-effort
// operation.
func (lim *Limiter) AddUpTo(now time.Time, n int64) int64 {
	lim.mu.Lock()
	defer lim.mu.Unlock()
	defer lim.curr.Sync(now)

	available := lim.capacity - lim.estimateCount(now)
	if n > available {
		n = available
	}
	lim.curr.AddCount(n)
	return available
}

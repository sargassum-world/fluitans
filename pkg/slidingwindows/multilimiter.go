// Package slidingwindows provides a multi-sliding-window meter for rate-limiting API requests
package slidingwindows

import (
	"sort"
	"sync"
	"time"
)

type Rate struct {
	AmountPerTime int64
	Time          time.Duration
}

type MultiLimiter struct {
	limiters map[string]*Limiter
	rates    map[string]Rate
	addLock  sync.Mutex
	// Keys are sorted by the duration of the corresponding limiter
	sortedKeys []string
}

func sortedKeys(limiters map[string]*Limiter) []string {
	type KeyTime struct {
		Key      string
		Duration time.Duration
	}
	// Get the limiter durations
	var keyTimes []KeyTime
	for k, limiter := range limiters {
		keyTimes = append(keyTimes, KeyTime{
			Key:      k,
			Duration: limiter.Duration(),
		})
	}
	// Sort keys by limiter duration
	sort.Slice(keyTimes, func(i, j int) bool {
		return keyTimes[i].Duration < keyTimes[j].Duration
	})
	sorted := make([]string, len(keyTimes))
	for i, kt := range keyTimes {
		sorted[i] = kt.Key
	}
	return sorted
}

func NewMultiLimiter(rates map[string]Rate) *MultiLimiter {
	limiters := make(map[string]*Limiter)
	for k, rateConfig := range rates {
		limiter, _ := NewLimiter(
			rateConfig.Time, rateConfig.AmountPerTime,
			func() (Window, StopFunc) {
				// We should not use a synchronizer which could update counts in the middle
				// of the TryAdd method, because that would change the TryAdd method from an
				// all-or-nothing operation to a best-effort operation.
				return NewLocalWindow()
			},
		)
		limiters[k] = limiter
	}
	return &MultiLimiter{
		limiters:   limiters,
		sortedKeys: sortedKeys(limiters),
		rates:      rates,
	}
}

func (ml *MultiLimiter) Rates() map[string]Rate {
	return ml.rates
}

type KeyFillRatio struct {
	Key       string
	FillRatio float64
}

func (ml *MultiLimiter) EstimateFillRatios(now time.Time) []KeyFillRatio {
	// Because these fill ratios are only approximate, we don't need to guarantee
	// accuracy against data races
	fillRatios := make(map[string]float64)
	for k, l := range ml.limiters {
		fillRatios[k] = l.EstimateFillRatio(now)
	}
	keyFillRatios := make([]KeyFillRatio, len(ml.limiters))
	for i, key := range ml.sortedKeys {
		keyFillRatios[i] = KeyFillRatio{
			Key:       key,
			FillRatio: fillRatios[key],
		}
	}
	return keyFillRatios
}

func (ml *MultiLimiter) EstimateWaitDuration(now time.Time, n int64) time.Duration {
	// Because the wait duration is only approximate, we don't need to guarantee
	// accuracy against data races
	var maxWaitDuration time.Duration = 0
	for _, key := range ml.sortedKeys {
		waitDuration := ml.limiters[key].EstimateWaitDuration(now, n)
		if waitDuration > maxWaitDuration {
			maxWaitDuration = waitDuration
		}
	}
	return maxWaitDuration
}

func (ml *MultiLimiter) MaybeAllowed(now time.Time, n int64) bool {
	// This method doesn't guarantee that TryAdd will actually succeed (it can only
	// guarantee that TryAdd will fail), so we don't need to worry about data races
	// which could cause TryAdd to fail, so we don't need a mutex lock here
	for _, key := range ml.sortedKeys {
		if !ml.limiters[key].MaybeAllowed(now, n) {
			return false
		}
	}
	return true
}

func (ml *MultiLimiter) TryAdd(now time.Time, n int64) bool {
	// We want a lock in this method so that clients don't starve each other by racing
	// to hit different limiters.
	ml.addLock.Lock()
	defer ml.addLock.Unlock()

	// Don't do anything if any limiter doesn't have enough room. Because we lock
	// all writes to any limiter in the addLock mutex and we don't expose the
	// individual limiters in our public interface and we perform the MaybeAllowed
	// check after locking write access to the limiters, the TryAdd method should be an
	// all-or-nothing operation - at least if we don't use a synchronizer which
	// could increase counts between these checks and the actual mutations.
	if !ml.MaybeAllowed(now, n) {
		return false
	}

	allAdded := true
	for _, key := range ml.sortedKeys {
		// Try to add to every limiter, but record whether any failed (indicating that
		// we hit a rate limit somewhere)
		allAdded = ml.limiters[key].TryAdd(now, n) && allAdded
	}
	return allAdded
}

func (ml *MultiLimiter) estimateRateLimiter(
	throttleDuration time.Duration,
) (index int, limiter *Limiter) {
	// Use the list of keys, sorted by the timescale of their rate limits, to search
	// for the limiter which is probably responsible for throttling
	for i, key := range ml.sortedKeys {
		if throttleDuration <= ml.rates[key].Time {
			// This limiter is probably responsible for the throttling because none of
			// the previous limiters (which are all at shorter timescales) could have been
			// responsible for the throttling
			return i, ml.limiters[key]
		}
	}
	lastIndex := len(ml.sortedKeys) - 1
	lastKey := ml.sortedKeys[lastIndex]
	return lastIndex, ml.limiters[lastKey]
}

func (ml *MultiLimiter) Throttled(now time.Time, waitSeconds float64) {
	waitDuration := time.Duration(waitSeconds * float64(time.Second))
	responsibleIndex, responsibleLimiter := ml.estimateRateLimiter(waitDuration)

	ml.addLock.Lock()
	defer ml.addLock.Unlock()

	amountAdded := responsibleLimiter.AddUpTo(now, responsibleLimiter.Capacity())
	// Update resource usage estimates for all longer-timescale limiters
	for i := responsibleIndex + 1; i < len(ml.sortedKeys); i++ {
		ml.limiters[ml.sortedKeys[i]].AddUpTo(now, amountAdded)
	}
}

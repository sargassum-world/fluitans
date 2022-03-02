package desec

import (
	"time"

	"github.com/sargassum-world/fluitans/pkg/slidingwindows"
)

func clampRatio(ratio float64) float64 {
	if ratio < 0 {
		return 0
	}

	if ratio > 1 {
		return 1
	}

	return ratio
}

func adjustAmount(amount int64, safetyMargin float64) int64 {
	return int64(float64(amount) * (1.0 - clampRatio(safetyMargin)))
}

func NewReadLimiter(safetyMargin float64) *slidingwindows.MultiLimiter {
	var maxPerSec int64 = 10
	var maxPerMin int64 = 50
	return slidingwindows.NewMultiLimiter(map[string]slidingwindows.Rate{
		"PerSec": {
			AmountPerTime: adjustAmount(maxPerSec, safetyMargin),
			Time:          time.Second,
		},
		"PerMin": {
			AmountPerTime: adjustAmount(maxPerMin, safetyMargin),
			Time:          time.Minute,
		},
	})
}

func NewDomainWriteLimiter(safetyMargin float64) *slidingwindows.MultiLimiter {
	var maxPerSec int64 = 10
	var maxPerMin int64 = 300
	var maxPerHour int64 = 1000
	return slidingwindows.NewMultiLimiter(map[string]slidingwindows.Rate{
		"PerSec": {
			AmountPerTime: adjustAmount(maxPerSec, safetyMargin),
			Time:          time.Second,
		},
		"PerMin": {
			AmountPerTime: adjustAmount(maxPerMin, safetyMargin),
			Time:          time.Minute,
		},
		"PerHour": {
			AmountPerTime: adjustAmount(maxPerHour, safetyMargin),
			Time:          time.Hour,
		},
	})
}

func NewRRSetWriteLimiter(safetyMargin float64) *slidingwindows.MultiLimiter {
	var maxPerSec int64 = 2
	var maxPerMin int64 = 15
	var maxPerHour int64 = 30
	var maxPerDay int64 = 300
	hoursPerDay := 24
	day := time.Duration(hoursPerDay) * time.Hour
	return slidingwindows.NewMultiLimiter(map[string]slidingwindows.Rate{
		"PerSec": {
			AmountPerTime: adjustAmount(maxPerSec, safetyMargin),
			Time:          time.Second,
		},
		"PerMin": {
			AmountPerTime: adjustAmount(maxPerMin, safetyMargin),
			Time:          time.Minute,
		},
		"PerHour": {
			AmountPerTime: adjustAmount(maxPerHour, safetyMargin),
			Time:          time.Hour,
		},
		"PerDay": {
			AmountPerTime: adjustAmount(maxPerDay, safetyMargin),
			Time:          day,
		},
	})
}

package client

import (
	"github.com/dgraph-io/ristretto"

	"github.com/sargassum-eco/fluitans/internal/env"
)

func GetEnvVarCacheConfig() (*ristretto.Config, error) {
	var defaultNumCounters int64 = 3e6 // default: 300k items, ~9 MB of counters
	numCounters, err := env.GetInt64("FLUITANS_CACHE_NUMCOUNTERS", defaultNumCounters)
	if err != nil {
		return nil, err
	}

	var defaultMaxCost int64 = 3e7 // default: up to 30 MB total with min cost weight of 1
	maxCost, err := env.GetInt64("FLUITANS_CACHE_MAXCOST", defaultMaxCost)
	if err != nil {
		return nil, err
	}

	var defaultBufferItems int64 = 64 // default: ristretto's recommended value
	bufferItems, err := env.GetInt64("FLUITANS_CACHE_BUFFERITEMS", defaultBufferItems)
	if err != nil {
		return nil, err
	}

	metrics, err := env.GetBool("FLUITANS_CACHE_METRICS")
	if err != nil {
		return nil, err
	}

	return &ristretto.Config{
		NumCounters: numCounters,
		MaxCost:     maxCost,
		BufferItems: bufferItems,
		Metrics:     metrics,
	}, nil
}

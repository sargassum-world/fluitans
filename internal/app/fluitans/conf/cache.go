package conf

import (
	"github.com/dgraph-io/ristretto"
	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/env"
)

func getCacheConfig() (c ristretto.Config, err error) {
	const defaultNumCounters = 3e6 // default: 300k items, ~9 MB of counters
	c.NumCounters, err = env.GetInt64("FLUITANS_CACHE_NUMCOUNTERS", defaultNumCounters)
	if err != nil {
		err = errors.Wrap(err, "couldn't make numCounters config")
		return
	}

	const defaultMaxCost = 3e7 // default: up to 30 MB total with min cost weight of 1
	c.MaxCost, err = env.GetInt64("FLUITANS_CACHE_MAXCOST", defaultMaxCost)
	if err != nil {
		err = errors.Wrap(err, "couldn't make maxCost config")
		return
	}

	const defaultBufferItems = 64 // default: ristretto's recommended value
	c.BufferItems, err = env.GetInt64("FLUITANS_CACHE_BUFFERITEMS", defaultBufferItems)
	if err != nil {
		err = errors.Wrap(err, "couldn't make bufferItems config")
		return
	}

	c.Metrics, err = env.GetBool("FLUITANS_CACHE_METRICS")
	if err != nil {
		err = errors.Wrap(err, "couldn't make metrics config")
		return
	}

	return
}
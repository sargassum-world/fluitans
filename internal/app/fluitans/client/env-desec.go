package client

import (
	"time"

	"github.com/sargassum-eco/fluitans/internal/env"
)

func GetEnvVarDesecAPISettings() (*DesecAPISettings, error) {
	// The TTL of cached read results, in units of seconds; negative numbers represent
	// an infinite TTL. All reads are cached, and cache entries below TTL will be used
	// instead of issuing extra API read requests. The cache will be consistent with
	// the API at an infinite TTL (the default TTL) if we promise to only modify DNS
	// records through Fluitans, and not independently through the deSEC server.
	rawReadCacheTTL, err := env.GetFloat32("FLUITANS_DESEC_READ_CACHE_TTL", -1)
	var readCacheTTL time.Duration = -1
	if rawReadCacheTTL >= 0 {
		durationReadCacheTTL := time.Duration(rawReadCacheTTL) * time.Second
		readCacheTTL = durationReadCacheTTL
	}
	if err != nil {
		return nil, err
	}

	// The write limiter fill ratio above which RRset writes, rather than being executed
	// immediately, will first be batched into groups based on the nearest rate limit
	var defaultWriteSoftQuota float32 = 0.34
	writeSoftQuota, err := env.GetFloat32(
		"FLUITANS_DESEC_WRITE_SOFT_QUOTA", defaultWriteSoftQuota,
	)
	if err != nil {
		return nil, err
	}

	if writeSoftQuota < 0 {
		writeSoftQuota = 0
	}
	if writeSoftQuota > 1 {
		writeSoftQuota = 1
	}

	return &DesecAPISettings{
		ReadCacheTTL:   readCacheTTL,
		WriteSoftQuota: writeSoftQuota,
	}, nil
}

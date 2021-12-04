package conf

import (
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/models"
	"github.com/sargassum-eco/fluitans/internal/env"
)

func getDesecReadCacheTTL() (time.Duration, error) {
	// The TTL of cached read results, in units of seconds; negative numbers represent
	// an infinite TTL. All reads are cached, and cache entries below TTL will be used
	// instead of issuing extra API read requests. The cache will be consistent with
	// the API at an infinite TTL (the default TTL) if we promise to only modify DNS
	// records through Fluitans, and not independently through the deSEC server.
	rawTTL, err := env.GetFloat32("FLUITANS_DESEC_READ_CACHE_TTL", -1)
	var ttl time.Duration = -1
	if rawTTL >= 0 {
		durationReadCacheTTL := time.Duration(rawTTL) * time.Second
		ttl = durationReadCacheTTL
	}
	if err != nil {
		return -1, err
	}

	return ttl, nil
}

func getDesecAPISettings() (*models.DesecAPISettings, error) {
	readCacheTTL, err := getDesecReadCacheTTL()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make readCacheTTL config")
	}

	// The write limiter fill ratio above which RRset writes, rather than being executed
	// immediately, will first be batched into groups based on the nearest rate limit
	var defaultWriteSoftQuota float32 = 0.34
	writeSoftQuota, err := env.GetFloat32(
		"FLUITANS_DESEC_WRITE_SOFT_QUOTA", defaultWriteSoftQuota,
	)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make writeSoftQuota config")
	}

	if writeSoftQuota < 0 {
		writeSoftQuota = 0
	}
	if writeSoftQuota > 1 {
		writeSoftQuota = 1
	}

	return &models.DesecAPISettings{
		ReadCacheTTL:   readCacheTTL,
		WriteSoftQuota: writeSoftQuota,
	}, nil
}

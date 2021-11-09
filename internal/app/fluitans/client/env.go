// Package client contains client code for external APIs
package client

import (
	"os"

	"github.com/dgraph-io/ristretto"

	"github.com/sargassum-eco/fluitans/internal/env"
)

func GetEnvVarController() (*Controller, error) {
	url, err := env.GetURL("FLUITANS_ZT_CONTROLLER_SERVER", "")
	if err != nil {
		return nil, err
	}

	if len(url.Scheme) == 0 {
		url.Scheme = "http"
	}
	url.Path = ""
	url.User = nil
	url.RawQuery = ""
	url.Fragment = ""

	networkCostWeight, err := env.GetFloat32("FLUITANS_ZT_CONTROLLER_NETWORKCOST", 1.0)
	if err != nil {
		return nil, err
	}

	authtoken := os.Getenv("FLUITANS_ZT_CONTROLLER_AUTHTOKEN")
	name := env.GetString("FLUITANS_ZT_CONTROLLER_NAME", url.Host)
	desc := env.GetString(
		"FLUITANS_ZT_CONTROLLER_DESC",
		"The default ZeroTier network controller specified in the environment variables.",
	)
	if len(url.String()) == 0 || len(authtoken) == 0 {
		return nil, nil
	}

	return &Controller{
		Server:            url.String(),
		Name:              name,
		Description:       desc,
		Authtoken:         authtoken,
		NetworkCostWeight: networkCostWeight,
	}, nil
}

func GetCacheConfig() (*ristretto.Config, error) {
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

func GetDomainName() string {
	return os.Getenv("FLUITANS_DOMAIN_NAME")
}

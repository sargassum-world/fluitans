package fluitans

import (
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/conf"
	"github.com/sargassum-eco/fluitans/pkg/desec"
	"github.com/sargassum-eco/fluitans/pkg/framework/log"
	"github.com/sargassum-eco/fluitans/pkg/slidingwindows"
)

func setupRateLimiters(
	config conf.Config,
) (map[string]*slidingwindows.MultiLimiter, error) {
	switch api := config.DNSServer.API; api {
	default:
		return nil, errors.Errorf("Unknown DNS Server API type: %s (allowed choices: desec)", api)
	case "desec":
		desecLimiters := map[string]*slidingwindows.MultiLimiter{
			client.DesecReadLimiterName:  desec.NewReadLimiter(0),
			client.DesecWriteLimiterName: desec.NewRRSetWriteLimiter(0),
		}
		return desecLimiters, nil
	}
}

func MakeAppGlobals(l log.Logger) (*client.Globals, error) {
	config, err := conf.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up application config")
	}

	cache, err := client.NewCache(config.Cache)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up client cache")
	}

	rateLimiters, err := setupRateLimiters(*config)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up rate limiters")
	}

	return &client.Globals{
		Config:       *config,
		Logger:       l,
		Cache:        cache,
		RateLimiters: rateLimiters,
		DNSDomain: &client.DNSDomain{
			Server:      config.DNSServer,
			APISettings: config.DesecAPI,
			DomainName:  config.DomainName,
			Cache:       cache,
			ReadLimiter: rateLimiters[client.DesecReadLimiterName],
		},
	}, nil
}

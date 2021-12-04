package fluitans

import (
	"io/fs"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/conf"
	"github.com/sargassum-eco/fluitans/pkg/desec"
	"github.com/sargassum-eco/fluitans/pkg/framework/fsutil"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/framework/template"
	"github.com/sargassum-eco/fluitans/pkg/slidingwindows"
	"github.com/sargassum-eco/fluitans/web"
)

type Globals struct {
	Template route.TemplateGlobals
	Static   route.StaticGlobals
}

func computeTemplateFingerprints() (*route.TemplateFingerprints, error) {
	layoutFiles, err := fsutil.ListFiles(web.TemplatesFS, template.FilterApp)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't load template layouts & partials")
	}

	pageFiles, err := fsutil.ListFiles(web.TemplatesFS, template.FilterPage)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't load template pages")
	}

	appFiles, err := fsutil.ListFiles(web.AppFS, template.FilterAsset)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't load app assets")
	}

	return route.ComputeTemplateFingerprints(
		layoutFiles, pageFiles, appFiles, web.TemplatesFS, web.AppFS,
	)
}

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

func makeTemplatedRouteEmbeds() route.Embeds {
	return route.Embeds{
		CSS: template.PreprocessCSS(template.EmbeddableAssets{
			"BundleEager": web.BundleEagerCSS,
		}),
		JS: template.PreprocessJS(template.EmbeddableAssets{
			"BundleEager": web.BundleEagerJS,
		}),
	}
}

func makeStaticGlobals() route.StaticGlobals {
	return route.StaticGlobals{
		FS: map[string]fs.FS{
			"Web":   web.StaticFS,
			"Fonts": web.FontsFS,
		},
		HFS: map[string]fs.FS{
			"Static": web.StaticHFS,
			"App":    web.AppHFS,
		},
	}
}

func makeAppGlobals() (*client.Globals, error) {
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

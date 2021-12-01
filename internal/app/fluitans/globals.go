package fluitans

import (
	"io/fs"
	"strings"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/templates"
	"github.com/sargassum-eco/fluitans/internal/fsutil"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
	"github.com/sargassum-eco/fluitans/pkg/desec"
	"github.com/sargassum-eco/fluitans/pkg/slidingwindows"
	"github.com/sargassum-eco/fluitans/web"
)

type Globals struct {
	Template route.TemplateGlobals
	Static   route.StaticGlobals
}

func computeTemplateFingerprints() (*route.TemplateFingerprints, error) {
	layoutFiles, err := fsutil.ListFiles(web.TemplatesFS, templates.FilterApp)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't load template layouts & partials")
	}

	pageFiles, err := fsutil.ListFiles(web.TemplatesFS, templates.FilterPage)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't load template pages")
	}

	appFiles, err := fsutil.ListFiles(web.AppFS, func(path string) bool {
		return strings.HasSuffix(path, ".min.css") || strings.HasSuffix(path, ".js")
	})
	if err != nil {
		return nil, errors.Wrap(err, "couldn't load app assets")
	}

	return route.ComputeTemplateFingerprints(
		layoutFiles, pageFiles, appFiles, web.TemplatesFS, web.AppFS,
	)
}

func setupCache() (*client.Cache, error) {
	cacheConfig, err := client.GetEnvVarCacheConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get client cache config from env vars")
	}

	cache, err := client.NewCache(cacheConfig)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make client cache")
	}

	return cache, nil
}

func setupRateLimiters() (map[string]*slidingwindows.MultiLimiter, error) {
	dnsServer, err := client.GetEnvVarDNSServer()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't get DNS server config from env vars")
	}

	switch api := dnsServer.API; api {
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

func makeTemplateEmbeds() template.Embeds {
	return template.Embeds{
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
	cache, err := setupCache()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up the global cache")
	}

	rateLimiters, err := setupRateLimiters()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up the global rate limiters")
	}

	return &client.Globals{
		Cache:        cache,
		RateLimiters: rateLimiters,
	}, nil
}

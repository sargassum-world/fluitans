// Package fluitans provides the Fluitans server.
package fluitans

import (
	"io/fs"
	"strings"

	"github.com/dgraph-io/ristretto"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/templates"
	"github.com/sargassum-eco/fluitans/internal/fsutil"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/web"
)

type Globals struct {
	Template route.TemplateGlobals
	Static   route.StaticGlobals
}

func computeGlobals() (*Globals, error) {
	layoutFiles, err := fsutil.ListFiles(web.TemplatesFS, templates.FilterApp)
	if err != nil {
		return nil, err
	}

	pageFiles, err := fsutil.ListFiles(web.TemplatesFS, templates.FilterPage)
	if err != nil {
		return nil, err
	}

	appFiles, err := fsutil.ListFiles(web.AppFS, func(path string) bool {
		return strings.HasSuffix(path, ".min.css") || strings.HasSuffix(path, ".js")
	})
	if err != nil {
		return nil, err
	}

	f, err := route.ComputeTemplateFingerprints(
		layoutFiles, pageFiles, appFiles, web.TemplatesFS, web.AppFS,
	)
	if err != nil {
		return nil, err
	}

	cacheConfig, err := client.GetCacheConfig()
	if err != nil {
		return nil, err
	}

	cache, err := ristretto.NewCache(cacheConfig)
	if err != nil {
		return nil, err
	}

	return &Globals{
		Template: route.TemplateGlobals{
			Embeds:               embeds,
			TemplateFingerprints: *f,
			Cache:                &client.Cache{Cache: cache},
		},
		Static: route.StaticGlobals{
			FS: map[string]fs.FS{
				"Web":   web.StaticFS,
				"Fonts": web.FontsFS,
			},
			HFS: map[string]fs.FS{
				"Static": web.StaticHFS,
				"App":    web.AppHFS,
			},
		},
	}, nil
}

func NewRenderer() *templates.TemplateRenderer {
	return templates.New(web.AppHFS.HashName, web.StaticHFS.HashName, web.TemplatesFS)
}

func RegisterRoutes(e route.EchoRouter) error {
	globals, err := computeGlobals()
	if err != nil {
		return err
	}

	err = route.RegisterTemplated(e, routes.TemplatedAssets, globals.Template)
	if err != nil {
		return err
	}

	err = route.RegisterStatic(e, routes.StaticAssets, globals.Static)
	if err != nil {
		return err
	}

	err = route.RegisterTemplated(e, routes.Pages, globals.Template)
	if err != nil {
		return err
	}

	return nil
}

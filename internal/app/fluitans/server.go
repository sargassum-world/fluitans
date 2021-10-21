// Package fluitans provides the Fluitans server.
package fluitans

import (
	"io/fs"
	"strings"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/templates"
	"github.com/sargassum-eco/fluitans/internal/fsutil"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/web"
)

// TODO: add a flag to specify reading all files from disk into a virtual FS memory,
// rather than using the compile-time virtual FS
// TODO: even better is if it acts as an overlay FS, so only the files to override need to be supplied

func computeGlobals() (*route.TemplateGlobals, *route.StaticGlobals, error) {
	layoutFiles, err := fsutil.ListFiles(web.TemplatesFS, templates.FilterApp)
	if err != nil {
		return nil, nil, err
	}

	pageFiles, err := fsutil.ListFiles(web.TemplatesFS, templates.FilterPage)
	if err != nil {
		return nil, nil, err
	}

	appFiles, err := fsutil.ListFiles(web.AppFS, func(path string) bool {
		return strings.HasSuffix(path, ".min.css") || strings.HasSuffix(path, ".js")
	})
	if err != nil {
		return nil, nil, err
	}

	f, err := route.ComputeTemplateFingerprints(
		layoutFiles, pageFiles, appFiles, web.TemplatesFS, web.AppFS,
	)
	if err != nil {
		return nil, nil, err
	}

	tg := route.TemplateGlobals{
		Embeds:               embeds,
		TemplateFingerprints: *f,
	}
	sg := route.StaticGlobals{
		FS: map[string]fs.FS{
			"Web":   web.StaticFS,
			"Fonts": web.FontsFS,
		},
		HFS: map[string]fs.FS{
			"Static": web.StaticHFS,
			"App":    web.AppHFS,
		},
	}
	return &tg, &sg, nil
}

func NewRenderer() *templates.TemplateRenderer {
	return templates.New(web.AppHFS.HashName, web.StaticHFS.HashName, web.TemplatesFS)
}

func RegisterRoutes(e route.EchoRouter) error {
	tg, sg, err := computeGlobals()
	if err != nil {
		return err
	}

	err = route.RegisterTemplated(e, routes.TemplatedAssets, *tg)
	if err != nil {
		return err
	}

	err = route.RegisterStatic(e, routes.StaticAssets, *sg)
	if err != nil {
		return err
	}

	err = route.RegisterTemplated(e, routes.Pages, *tg)
	if err != nil {
		return err
	}

	return nil
}

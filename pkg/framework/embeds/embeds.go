// Package embeds provides support for embedding all web app files
package embeds

import (
	"html/template"
	"io/fs"

	"github.com/benbjohnson/hashfs"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework/fsutil"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	tp "github.com/sargassum-eco/fluitans/pkg/framework/template"
)

type Embeds struct {
	StaticFS    fs.FS
	StaticHFS   *hashfs.FS
	TemplatesFS fs.FS
	AppFS       fs.FS
	AppHFS      *hashfs.FS
	FontsFS     fs.FS
	Inlines     route.Inlines
}

func (e Embeds) ComputeTemplateFingerprints() (*route.TemplateFingerprints, error) {
	layoutFiles, err := fsutil.ListFiles(e.TemplatesFS, tp.FilterApp)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't load template layouts & partials")
	}

	pageFiles, err := fsutil.ListFiles(e.TemplatesFS, tp.FilterPage)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't load template pages")
	}

	appFiles, err := fsutil.ListFiles(e.AppFS, tp.FilterAsset)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't load app assets")
	}

	return route.ComputeTemplateFingerprints(
		layoutFiles, pageFiles, appFiles, e.TemplatesFS, e.AppFS,
	)
}

func (e Embeds) MakeStaticGlobals() route.StaticGlobals {
	return route.StaticGlobals{
		FS: map[string]fs.FS{
			"Web":   e.StaticFS,
			"Fonts": e.FontsFS,
		},
		HFS: map[string]fs.FS{
			"Static": e.StaticHFS,
			"App":    e.AppHFS,
		},
	}
}

func (e Embeds) NewTemplateRenderer(functions ...template.FuncMap) *tp.TemplateRenderer {
	return tp.NewTemplateRenderer(
		e.TemplatesFS,
		append(functions, tp.FuncMap(e.AppHFS.HashName, e.StaticHFS.HashName))...,
	)
}

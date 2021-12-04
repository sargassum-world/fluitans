// Package framework provides a reusable framework for Fluitans-style web apps
package framework

import (
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework/embeds"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

type Globals struct {
	Template route.TemplateGlobals
	Static   route.StaticGlobals
}

func NewGlobals(e embeds.Embeds, appGlobals interface{}) (*Globals, error) {
	f, err := e.ComputeTemplateFingerprints()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't compute template fingerprints")
	}

	return &Globals{
		Template: route.TemplateGlobals{
			Inlines:              e.Inlines,
			TemplateFingerprints: *f,
			App:                  appGlobals,
		},
		Static: e.MakeStaticGlobals(),
	}, nil
}

func (g *Globals) RegisterRoutes(
	e route.EchoRouter,
	templatedAssets []route.Templated, staticAssets []route.Static,
	templatedPages []route.Templated,
) error {
	if err := route.RegisterTemplated(e, templatedAssets, g.Template); err != nil {
		return errors.Wrap(err, "couldn't register templated assets")
	}

	if err := route.RegisterStatic(e, staticAssets, g.Static); err != nil {
		return errors.Wrap(err, "couldn't register static assets")
	}

	if err := route.RegisterTemplated(e, templatedPages, g.Template); err != nil {
		return errors.Wrap(err, "couldn't register templated routes")
	}

	return nil
}

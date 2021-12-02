// Package fluitans provides the Fluitans server.
package fluitans

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/templates"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
	"github.com/sargassum-eco/fluitans/web"
)

func NewRenderer() *templates.TemplateRenderer {
	return templates.New(web.AppHFS.HashName, web.StaticHFS.HashName, web.TemplatesFS)
}

func RegisterRoutes(e route.EchoRouter) (*Globals, error) {
	f, err := computeTemplateFingerprints()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't compute template fingerprints")
	}

	ag, err := makeAppGlobals()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make app globals")
	}

	g := &Globals{
		Template: route.TemplateGlobals{
			Embeds:               makeTemplateEmbeds(),
			TemplateFingerprints: *f,
			App:                  ag,
		},
		Static: makeStaticGlobals(),
	}

	err = route.RegisterTemplated(e, routes.TemplatedAssets, g.Template)
	if err != nil {
		return g, errors.Wrap(err, "couldn't register templated assets")
	}

	err = route.RegisterStatic(e, routes.StaticAssets, g.Static)
	if err != nil {
		return g, errors.Wrap(err, "couldn't register static assets")
	}

	err = route.RegisterTemplated(e, routes.Pages, g.Template)
	if err != nil {
		return g, errors.Wrap(err, "couldn't register templated routes")
	}

	return g, nil
}

func NewHTTPErrorHandler(g *Globals) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if herr, ok := err.(*echo.HTTPError); ok {
			code = herr.Code
		}
		perr := c.Render(code, "httperr.page.tmpl", struct {
			Meta   template.Meta
			Embeds template.Embeds
			Data   int
		}{
			Meta: template.Meta{
				Path:       c.Request().URL.Path,
				DomainName: client.GetEnvVarDomainName(),
			},
			Embeds: g.Template.Embeds,
			Data:   code,
		})
		if perr != nil {
			c.Logger().Error(err)
		}
		c.Logger().Error(err)
	}
}

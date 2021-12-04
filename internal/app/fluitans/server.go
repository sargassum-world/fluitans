// Package fluitans provides the Fluitans server.
package fluitans

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/templates"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/workers"
	"github.com/sargassum-eco/fluitans/internal/log"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
	"github.com/sargassum-eco/fluitans/web"
)

func NewRenderer() *templates.TemplateRenderer {
	return templates.New(web.AppHFS.HashName, web.StaticHFS.HashName, web.TemplatesFS)
}

func MakeGlobals() (*Globals, error) {
	f, err := computeTemplateFingerprints()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't compute template fingerprints")
	}

	ag, err := makeAppGlobals()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make app globals")
	}

	return &Globals{
		Template: route.TemplateGlobals{
			Embeds:               makeTemplateEmbeds(),
			TemplateFingerprints: *f,
			App:                  ag,
		},
		Static: makeStaticGlobals(),
	}, nil
}

func RegisterRoutes(e route.EchoRouter, g *Globals) error {
	if err := route.RegisterTemplated(e, routes.TemplatedAssets, g.Template); err != nil {
		return errors.Wrap(err, "couldn't register templated assets")
	}

	if err := route.RegisterStatic(e, routes.StaticAssets, g.Static); err != nil {
		return errors.Wrap(err, "couldn't register static assets")
	}

	if err := route.RegisterTemplated(e, routes.Pages, g.Template); err != nil {
		return errors.Wrap(err, "couldn't register templated routes")
	}

	return nil
}

func NewHTTPErrorHandler(g *Globals) (func(err error, c echo.Context), error) {
	switch app := g.Template.App.(type) {
	default:
		return nil, fmt.Errorf("app globals are of unexpected type %T", g.Template.App)
	case *client.Globals:
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
					DomainName: app.Config.DomainName,
				},
				Embeds: g.Template.Embeds,
				Data:   code,
			})
			if perr != nil {
				c.Logger().Error(err)
			}
			c.Logger().Error(err)
		}, nil
	}
}

func LaunchBackgroundWorkers(g *Globals, l log.Logger) error {
	switch app := g.Template.App.(type) {
	default:
		return fmt.Errorf("app globals are of unexpected type %T", g.Template.App)
	case *client.Globals:
		go workers.PrescanZerotierControllers(app, l)
		go workers.PrefetchDNSRecords(app, l)
		go workers.TestWriteLimiter(app)
	}
	return nil
}

func PrepareServer(e *echo.Echo) error {
	e.Renderer = NewRenderer()

	g, err := MakeGlobals()
	if err != nil {
		return err
	}

	errorHandler, err := NewHTTPErrorHandler(g)
	if err != nil {
		return err
	}
	e.HTTPErrorHandler = errorHandler

	if err = RegisterRoutes(e, g); err != nil {
		return err
	}

	if err = LaunchBackgroundWorkers(g, e.Logger); err != nil {
		return err
	}

	return nil
}

// Package fluitans provides the Fluitans server.
package fluitans

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/tmplfunc"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/workers"
	"github.com/sargassum-eco/fluitans/pkg/framework/log"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/framework/template"
	"github.com/sargassum-eco/fluitans/web"
)

func NewRenderer() *template.TemplateRenderer {
	return template.NewRenderer(
		web.TemplatesFS,
		template.FuncMap(web.AppHFS.HashName, web.StaticHFS.HashName),
		tmplfunc.FuncMap,
	)
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
			Embeds:               makeTemplatedRouteEmbeds(),
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
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if herr, ok := err.(*echo.HTTPError); ok {
			code = herr.Code
		}
		perr := c.Render(
			code, "httperr.page.tmpl", route.MakeRenderData(c, g.Template, code),
		)
		if perr != nil {
			c.Logger().Error(err)
		}
		c.Logger().Error(err)
	}, nil
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

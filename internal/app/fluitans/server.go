// Package fluitans provides the Fluitans server.
package fluitans

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/tmplfunc"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/workers"
	"github.com/sargassum-eco/fluitans/pkg/framework"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/web"
)

func NewHTTPErrorHandler(g *framework.Globals) (func(err error, c echo.Context), error) {
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

func RegisterRoutes(g *framework.Globals, e route.EchoRouter) error {
	return g.RegisterRoutes(e, routes.TemplatedAssets, routes.StaticAssets, routes.Pages)
}

func LaunchBackgroundWorkers(ag *client.Globals) {
	go workers.PrescanZerotierControllers(ag)
	go workers.PrefetchDNSRecords(ag)
	go workers.TestWriteLimiter(ag)
}

func PrepareServer(e *echo.Echo) error {
	embeds := web.NewEmbeds()
	e.Renderer = embeds.NewTemplateRenderer(tmplfunc.FuncMap)

	// Globals
	ag, err := MakeAppGlobals(e.Logger)
	if err != nil {
		return errors.Wrap(err, "couldn't make app globals")
	}

	g, err := framework.NewGlobals(embeds, ag)
	if err != nil {
		return errors.Wrap(err, "couldn't make server globals")
	}

	// Routes
	if err = RegisterRoutes(g, e); err != nil {
		return errors.Wrap(err, "couldn't register routes")
	}

	// Error Handling
	errorHandler, err := NewHTTPErrorHandler(g)
	if err != nil {
		return errors.Wrap(err, "couldn't register HTTP error handler")
	}
	e.HTTPErrorHandler = errorHandler

	// Background Workers
	LaunchBackgroundWorkers(ag)

	return nil
}

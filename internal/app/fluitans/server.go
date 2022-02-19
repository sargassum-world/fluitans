// Package fluitans provides the Fluitans server.
package fluitans

import (
	"github.com/labstack/echo-contrib/session"
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

func RegisterRoutes(g *framework.Globals, e route.EchoRouter) error {
	return g.RegisterRoutes(e, routes.TemplatedAssets, routes.StaticAssets, routes.Pages)
}

func LaunchBackgroundWorkers(ag *client.Globals) {
	go workers.PrescanZerotierControllers(ag.Clients.ZTControllers)
	go workers.PrefetchZerotierNetworks(ag.Clients.Zerotier, ag.Clients.ZTControllers)
	go workers.PrefetchDNSRecords(ag.Clients.Desec)
	// go workers.TestWriteLimiter(ag.Clients.Desec)
}

func PrepareServer(e *echo.Echo) error {
	embeds := web.NewEmbeds()
	e.Renderer = embeds.NewTemplateRenderer(tmplfunc.FuncMap)

	// Globals
	ag, err := client.NewGlobals(e.Logger)
	if err != nil {
		return errors.Wrap(err, "couldn't make app globals")
	}
	g, err := framework.NewGlobals(embeds, ag)
	if err != nil {
		return errors.Wrap(err, "couldn't make server globals")
	}

	// Session Management
	e.Use(session.Middleware(ag.Clients.Sessions.Store))

	// Routes
	if err = RegisterRoutes(g, e); err != nil {
		return errors.Wrap(err, "couldn't register routes")
	}

	// Error Handling
	e.HTTPErrorHandler, err = NewHTTPErrorHandler(g.Template, ag.Clients.Sessions)
	if err != nil {
		return errors.Wrap(err, "couldn't register HTTP error handler")
	}

	// Background Workers
	LaunchBackgroundWorkers(ag)

	return nil
}

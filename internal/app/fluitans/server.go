// Package fluitans provides the Fluitans server.
package fluitans

import (
	"github.com/Masterminds/sprig/v3"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/assets"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/tmplfunc"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/workers"
	"github.com/sargassum-eco/fluitans/pkg/framework"
	"github.com/sargassum-eco/fluitans/web"
)

func LaunchBackgroundWorkers(clients *client.Clients) {
	go workers.PrescanZerotierControllers(clients.ZTControllers)
	go workers.PrefetchZerotierNetworks(clients.Zerotier, clients.ZTControllers)
	go workers.PrefetchDNSRecords(clients.Desec)
	// go workers.TestWriteLimiter(clients.Desec)
}

func PrepareServer(e *echo.Echo) error {
	embeds := web.NewEmbeds()
	inlines := web.NewInlines()
	tr, err := framework.NewTemplateRenderer(
		embeds, inlines, sprig.FuncMap(), tmplfunc.FuncMap(
			tmplfunc.NewHashedNamers(assets.AppURLPrefix, assets.StaticURLPrefix, embeds),
		),
	)
	if err != nil {
		return errors.Wrap(err, "couldn't make template renderer")
	}
	e.Renderer = tr.GetEchoRenderer()

	// Globals
	g, err := client.NewGlobals(e.Logger)
	if err != nil {
		return errors.Wrap(err, "couldn't make app globals")
	}

	// Middlewares & Route Handlers
	e.Use(g.Clients.Sessions.NewCSRFMiddleware(
		csrf.ErrorHandler(NewCSRFErrorHandler(tr, e.Logger, g.Clients.Sessions)),
	))

	e.HTTPErrorHandler = NewHTTPErrorHandler(tr, g.Clients.Sessions)
	routes.NewService(tr, g.Clients).Register(e, embeds)

	LaunchBackgroundWorkers(g.Clients)
	return nil
}

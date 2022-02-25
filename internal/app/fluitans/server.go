// Package fluitans provides the Fluitans server.
package fluitans

import (
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/assets"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/tmplfunc"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/workers"
	"github.com/sargassum-eco/fluitans/pkg/framework"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/web"
)

func LaunchBackgroundWorkers(ag *client.Globals) {
	go workers.PrescanZerotierControllers(ag.Clients.ZTControllers)
	go workers.PrefetchZerotierNetworks(ag.Clients.Zerotier, ag.Clients.ZTControllers)
	go workers.PrefetchDNSRecords(ag.Clients.Desec)
	// go workers.TestWriteLimiter(ag.Clients.Desec)
}

func PrepareServer(e *echo.Echo) error {
	embeds := web.NewEmbeds()
	r := embeds.NewTemplateRenderer(tmplfunc.FuncMap)
	e.Renderer = r

	// Globals
	ag, err := client.NewGlobals(e.Logger)
	if err != nil {
		return errors.Wrap(err, "couldn't make app globals")
	}
	g, err := framework.NewGlobals(embeds)
	if err != nil {
		return errors.Wrap(err, "couldn't make server globals")
	}

	// CSRF Defense
	e.Use(ag.Clients.Sessions.NewCSRFMiddleware(
		csrf.ErrorHandler(NewCSRFErrorHandler(g.Template, r, e.Logger, ag.Clients.Sessions)),
	))

	// Routes
	assets.RegisterStatic(e, embeds)
	if err := route.RegisterTemplated(
		e, assets.NewTemplatedService().Routes(), g.Template,
	); err != nil {
		return errors.Wrap(err, "couldn't register templated assets")
	}
	if err := route.RegisterTemplated(
		e, routes.NewService(ag.Clients).Routes(), g.Template,
	); err != nil {
		return errors.Wrap(err, "couldn't register templated routes")
	}

	// Error Handling
	e.HTTPErrorHandler = NewHTTPErrorHandler(g.Template, ag.Clients.Sessions)

	// Background Workers
	LaunchBackgroundWorkers(ag)

	return nil
}

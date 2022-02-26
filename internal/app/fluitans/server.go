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
	"github.com/sargassum-eco/fluitans/internal/middleware"
	"github.com/sargassum-eco/fluitans/pkg/godest"
	"github.com/sargassum-eco/fluitans/web"
)

type Server struct {
	Embeds   godest.Embeds
	Inlines  web.Inlines
	Renderer godest.TemplateRenderer
	Globals  *client.Globals
	Handlers *routes.Handlers
}

func NewServer(e *echo.Echo) (s *Server, err error) {
	s = &Server{}
	s.Embeds = web.NewEmbeds()
	s.Inlines = web.NewInlines()
	s.Renderer, err = godest.NewTemplateRenderer(
		s.Embeds, s.Inlines, sprig.FuncMap(), tmplfunc.FuncMap(
			tmplfunc.NewHashedNamers(assets.AppURLPrefix, assets.StaticURLPrefix, s.Embeds),
		),
	)
	if err != nil {
		s = nil
		err = errors.Wrap(err, "couldn't make template renderer")
		return
	}

	s.Globals, err = client.NewGlobals(e.Logger)
	if err != nil {
		s = nil
		err = errors.Wrap(err, "couldn't make app globals")
		return
	}

	s.Handlers = routes.New(s.Renderer, s.Globals.Clients)
	return
}

func (s *Server) Register(e *echo.Echo) {
	// Middleware
	e.Use(s.Globals.Clients.Sessions.NewCSRFMiddleware(
		csrf.ErrorHandler(NewCSRFErrorHandler(s.Renderer, e.Logger, s.Globals.Clients.Sessions)),
	))
	e.Use(middleware.RequireContentTypes(echo.MIMEApplicationForm))

	// Handlers
	e.HTTPErrorHandler = NewHTTPErrorHandler(s.Renderer, s.Globals.Clients.Sessions)
	s.Handlers.Register(e, s.Embeds)
}

func (s *Server) LaunchBackgroundWorkers() {
	go workers.PrescanZerotierControllers(s.Globals.Clients.ZTControllers)
	go workers.PrefetchZerotierNetworks(s.Globals.Clients.Zerotier, s.Globals.Clients.ZTControllers)
	go workers.PrefetchDNSRecords(s.Globals.Clients.Desec)
	// go workers.TestWriteLimiter(s.Globals.Clients.Desec)
}

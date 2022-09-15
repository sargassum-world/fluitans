// Package fluitans provides the Fluitans server.
package fluitans

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Masterminds/sprig/v3"
	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest"
	gmw "github.com/sargassum-world/godest/middleware"
	"github.com/unrolled/secure"
	"github.com/unrolled/secure/cspbuilder"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/assets"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/tmplfunc"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/workers"
	"github.com/sargassum-world/fluitans/web"
)

type Server struct {
	Globals  *client.Globals
	Embeds   godest.Embeds
	Inlines  godest.Inlines
	Renderer godest.TemplateRenderer
	Handlers *routes.Handlers
}

func NewServer(logger godest.Logger) (s *Server, err error) {
	s = &Server{}
	s.Globals, err = client.NewGlobals(logger)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make app globals")
	}

	s.Embeds = web.NewEmbeds()
	s.Inlines = web.NewInlines()
	s.Renderer, err = godest.NewTemplateRenderer(
		s.Embeds, s.Inlines, sprig.FuncMap(), tmplfunc.FuncMap(
			tmplfunc.NewHashedNamers(assets.AppURLPrefix, assets.StaticURLPrefix, s.Embeds),
			s.Globals.TSSigner.Sign,
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make template renderer")
	}

	s.Handlers = routes.New(s.Renderer, s.Globals)
	return s, nil
}

// Echo

func (s *Server) configureLogging(e *echo.Echo) {
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${remote_ip} ${method} ${uri} (${bytes_in}b) => " +
			"(${bytes_out}b after ${latency_human}) ${status} ${error}\n",
	}))
	e.HideBanner = true
	e.HidePort = true
	e.Logger.SetLevel(log.INFO) // TODO: set level via env var
}

func (s *Server) configureHeaders(e *echo.Echo) error {
	cspBuilder := cspbuilder.Builder{
		Directives: map[string][]string{
			cspbuilder.DefaultSrc: {"'self'"},
			cspbuilder.ScriptSrc: append(
				// Warning: script-src 'self' may not be safe to use if we're hosting user-uploaded content.
				// Then we'll need to provide hashes for scripts & styles we include by URL, and we'll need
				// to add the SRI integrity attribute to the tags including those files; however, it's
				// unclear how well-supported they are by browsers.
				[]string{"'self'", "'unsafe-inline'"},
				s.Inlines.ComputeJSHashesForCSP()...,
			),
			cspbuilder.StyleSrc: append(
				[]string{
					"'self'",
					"'unsafe-inline'",
					// Note: Turbo Drive tries to install a style tag for its progress bar, which leads to a CSP
					// error. We add a hash for it here, assuming ProgressBar.animationDuration == 300:
					"'sha512-rVca7GmrbBAUUoTnu9V9a6ZR4WAZdxFUnrsg3B+1zEsES4K6q7EW02LIXdYmE5aofGOwLySKKtOafC0hq892BA=='",
				},
				s.Inlines.ComputeCSSHashesForCSP()...,
			),
			cspbuilder.ObjectSrc:      {"'none'"},
			cspbuilder.ChildSrc:       {"'self'"},
			cspbuilder.BaseURI:        {"'none'"},
			cspbuilder.FormAction:     {"'self'"},
			cspbuilder.FrameAncestors: {"'none'"},
			// TODO: add HTTPS-related settings for CSP, including upgrade-insecure-requests
		},
	}
	csp, err := cspBuilder.Build()
	if err != nil {
		return errors.Wrap(err, "couldn't build content security policy")
	}

	e.Use(echo.WrapMiddleware(secure.New(secure.Options{
		// TODO: add HTTPS options
		FrameDeny:               true,
		ContentTypeNosniff:      true,
		ContentSecurityPolicy:   csp,
		ReferrerPolicy:          "no-referrer",
		CrossOriginOpenerPolicy: "same-origin",
	}).Handler))
	e.Use(echo.WrapMiddleware(gmw.SetCORP("same-site")))
	e.Use(echo.WrapMiddleware(gmw.SetCOEP("require-corp")))
	return nil
}

func (s *Server) Register(e *echo.Echo) error {
	e.Use(middleware.Recover())
	s.configureLogging(e)
	if err := s.configureHeaders(e); err != nil {
		return errors.Wrap(err, "couldn't configure http headers")
	}

	// Compression Middleware
	e.Use(middleware.Decompress())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: s.Globals.Config.HTTP.GzipLevel,
	}))

	// Other Middleware
	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(echo.WrapMiddleware(s.Globals.Sessions.NewCSRFMiddleware(
		csrf.ErrorHandler(NewCSRFErrorHandler(s.Renderer, e.Logger, s.Globals.Sessions)),
	)))
	e.Use(gmw.RequireContentTypes(echo.MIMEApplicationForm))
	// TODO: enable Prometheus and rate-limiting

	// Handlers
	e.HTTPErrorHandler = NewHTTPErrorHandler(s.Renderer, s.Globals.Sessions)
	s.Handlers.Register(e, s.Globals.TSBroker, s.Embeds)
	return nil
}

// Running

func (s *Server) runWorkersInContext(ctx context.Context) error {
	eg, _ := errgroup.WithContext(ctx) // Workers run independently, so we don't need egctx
	eg.Go(func() error {
		if err := workers.PrescanZerotierControllers(
			ctx, s.Globals.ZTControllers,
		); err != nil && err != context.Canceled {
			s.Globals.Logger.Error(errors.Wrap(err, "couldn't prescan zerotier controllers"))
		}
		return nil
	})
	eg.Go(func() error {
		if err := workers.PrefetchZerotierNetworks(
			ctx, s.Globals.Zerotier, s.Globals.ZTControllers,
		); err != nil && err != context.Canceled {
			s.Globals.Logger.Error(errors.Wrap(err, "couldn't prefetch zerotier networks"))
		}
		return nil
	})
	eg.Go(func() error {
		if err := workers.PrefetchDNSRecords(
			ctx, s.Globals.Desec,
		); err != nil && err != context.Canceled {
			s.Globals.Logger.Error(errors.Wrap(err, "couldn't prefetch dns records"))
		}
		return nil
	})
	// TODO: add worker to batch DNS record writes when needed
	eg.Go(func() error {
		if err := s.Globals.TSBroker.Serve(ctx); err != nil && err != context.Canceled {
			s.Globals.Logger.Error(errors.Wrap(
				err, "turbo streams broker encountered error while serving",
			))
		}
		return nil
	})
	return eg.Wait()
}

const port = 3000 // TODO: configure this with env var

func (s *Server) Run(e *echo.Echo) error {
	s.Globals.Logger.Info("starting fluitans server")
	// The echo http server can't be canceled by context cancelation, so the API shouldn't promise to
	// stop blocking execution on context cancelation - so we use the background context here. The
	// http server should instead be stopped gracefully by calling the Shutdown method, or forcefully
	// by calling the Close method.
	eg, egctx := errgroup.WithContext(context.Background())
	eg.Go(func() error {
		s.Globals.Logger.Info("starting background workers")
		if err := s.runWorkersInContext(egctx); err != nil {
			s.Globals.Logger.Error(errors.Wrap(
				err, "background worker encountered error",
			))
		}
		return nil
	})
	eg.Go(func() error {
		address := fmt.Sprintf(":%d", port)
		s.Globals.Logger.Infof("starting http server on %s", address)
		return e.Start(address)
	})
	if err := eg.Wait(); err != http.ErrServerClosed {
		return errors.Wrap(err, "http server encountered error")
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context, e *echo.Echo) (err error) {
	if errEcho := e.Shutdown(ctx); errEcho != nil {
		s.Globals.Logger.Error(errors.Wrap(errEcho, "couldn't shut down http server"))
		err = errEcho
	}
	return err
}

func (s *Server) Close(e *echo.Echo) error {
	return errors.Wrap(e.Close(), "http server encountered error when closing an underlying listener")
}

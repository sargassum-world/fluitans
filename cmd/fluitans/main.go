package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans"
	"github.com/sargassum-eco/fluitans/internal/templates"
)

func main() {
	e := echo.New()

	// Middleware
	// TODO: add logging configuration
	e.Use(middleware.Logger())
	// TODO: add recovery configuration
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 6,
	}))
	e.Use(middleware.Decompress())
	// TODO: enable CORS, CSRF, auth, Prometheus, rate-limiting, and security
	e.Pre(middleware.RemoveTrailingSlash())

	// Renderer
	renderer := templates.New()
	e.Renderer = renderer

	// Handlers
	fluitans.RegisterHandlers(e)
	fluitans.RegisterStatics(e)

	// Start server
	e.Logger.Fatal(e.Start(":3000"))
}

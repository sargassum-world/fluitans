package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/handlers"
	"github.com/sargassum-eco/fluitans/internal/templates"
)

const (
	gzipLevel = 6
	port      = 3000
)

func main() {
	e := echo.New()

	// Middleware
	// TODO: add logging configuration
	e.Use(middleware.Logger())
	// TODO: add recovery configuration
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: gzipLevel,
	}))
	e.Use(middleware.Decompress())
	// TODO: enable CORS, CSRF, auth, Prometheus, rate-limiting, and security
	e.Pre(middleware.RemoveTrailingSlash())

	// Renderer
	renderer := templates.New()
	e.Renderer = renderer

	// Handlers
	handlers.RegisterPages(e)
	handlers.RegisterAssets(e)

	// Start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

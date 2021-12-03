package main

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans"
)

const (
	gzipLevel = 6
	port      = 3000
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "${remote_ip} ${method} ${uri} (${bytes_in}b) => (${bytes_out}b after ${latency_human}) ${status} ${error}\n",
	}))
	e.Logger.SetLevel(log.WARN)
	e.Use(middleware.Recover())
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: gzipLevel,
	}))
	e.Use(middleware.Decompress())
	// TODO: enable CORS, CSRF, auth, Prometheus, rate-limiting, and security
	e.Pre(middleware.RemoveTrailingSlash())

	// Prepare server
	if err := fluitans.PrepareServer(e); err != nil {
		fmt.Printf("%+v\n", err)
		panic(err)
	}

	// Start server
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", port)))
}

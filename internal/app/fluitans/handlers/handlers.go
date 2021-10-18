// Package handlers contains the route handlers for the Fluitans server.
package handlers

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

// Utilities

func wrapAssetCacheControl(h echo.HandlerFunc, age int) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", age))
		return h(c)
	}
}

// EchoRouter is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration.
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

func RegisterPages(e EchoRouter) {
	e.GET("/", home)
	e.GET("/login", login)
	e.GET("/hello/:name", helloName)
}

// Static Handlers

func RegisterAssets(e EchoRouter) {
	// TODO: set up live reloading of the server when files change in web/static or web/app/public/build or the go files
	const hour = 60 * 60
	const day = 24 * hour
	const week = 7 * day
	const year = 365 * day

	e.GET("/app/app.webmanifest", wrapAssetCacheControl(webmanifest, week))
	e.GET("/favicon.ico", wrapAssetCacheControl(favicon, week))
	e.GET("/fonts/*", wrapAssetCacheControl(fonts, year))

	// HashFS handlers already set Cache-Control headers
	e.GET("/static/*", static)
	e.GET("/app/*", app)
}

// Package fluitans contains the application code for the Fluitans server.
package fluitans

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/web"
)

type Meta struct {
	Description string
	Path string
}

type Assets struct {
	BundleEagerJS  template.JS
	BundleEagerCSS template.CSS
}

var assets = Assets{
	BundleEagerJS:  template.JS(web.BundleEagerJS),
	BundleEagerCSS: template.CSS(web.BundleEagerCSS),
}

// Templated Handlers

func home(c echo.Context) error {
	return c.Render(http.StatusOK, "home.page.tmpl", struct {
		Meta Meta
		Assets Assets
	}{
		Meta: Meta{
			Description: "An application for managing Sargassum networks, domain names, and organizations.",
			Path: c.Request().URL.Path,
		},
		Assets: assets,
	})
}

func login(c echo.Context) error {
	return c.Render(http.StatusOK, "login.page.tmpl", struct {
		Meta Meta
		Assets Assets
	}{
		Meta: Meta{
			Description: "Log in to Fluitans.",
			Path: c.Request().URL.Path,
		},
		Assets: assets,
	})
}

func helloName(c echo.Context) error {
	name := c.Param("name")
	return c.Render(http.StatusOK, "hello.page.tmpl", struct {
		Meta Meta
		Assets Assets
		Name   string
	}{
		Meta: Meta{
			Description: fmt.Sprintf("A greeting for %s.", name),
			Path: c.Request().URL.Path,
		},
		Assets: assets,
		Name: name,
	})
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

func RegisterHandlers(e EchoRouter) {
	e.GET("/", home)
	e.GET("/login", login)
	e.GET("/hello/:name", helloName)
}

// Static Handlers

func WrapCacheControl(h echo.HandlerFunc, age int) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", age))
		return h(c)
	}
}

func RegisterStatics(e EchoRouter) {
	// TODO: allow toggling whether to read from the filesystem (in development mode) or to use the static FS
	// TODO: then set up live reloading of the server
	// Routes to serve favicons
	faviconHandler := http.FileServer(http.FS(web.StaticFS))
	e.GET("/favicon.ico", WrapCacheControl(echo.WrapHandler(faviconHandler), 60*60*24*7))

	// TODO: add cache-busting
	// /static route to serve static files
	staticHandler := http.FileServer(http.FS(web.StaticFS))
	e.GET("/static/*", WrapCacheControl(echo.WrapHandler(http.StripPrefix("/static/", staticHandler)), 60*60*24))

	// /app route to statically serve webapp build artifacts
	appHandler := http.FileServer(http.FS(web.AppFS))
	e.GET("/app/*", WrapCacheControl(echo.WrapHandler(http.StripPrefix("/app/", appHandler)), 60*60))
}

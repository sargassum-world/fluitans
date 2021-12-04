package routes

import (
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/templates"
	"github.com/sargassum-eco/fluitans/internal/caching"
	"github.com/sargassum-eco/fluitans/internal/route"
)

const (
	hour = 60 * 60
	day  = 24 * hour
	week = 7 * day
	year = 365 * day
)

var TemplatedAssets = []route.Templated{
	{
		Path:         "/app/app.webmanifest",
		Method:       http.MethodGet,
		HandlerMaker: webmanifest,
		Templates:    []string{"app.webmanifest.tmpl"},
	},
}

var StaticAssets = []route.Static{
	{
		Path:         "/favicon.ico",
		HandlerMaker: favicon,
	},
	{
		Path:         "/fonts/*",
		HandlerMaker: fonts,
	},
	{
		Path:         "/static/*",
		HandlerMaker: static,
	},
	{
		Path:         "/app/*",
		HandlerMaker: app,
	},
}

func webmanifest(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "app.webmanifest.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	return func(c echo.Context) error {
		// Handle Etag
		if noContent, err := templates.ProcessEtag(c, tte, ""); noContent {
			return err
		}

		// Render template
		c.Response().Header().Set(echo.HeaderContentType, "application/manifest+json")
		return c.Render(http.StatusOK, t, struct{}{})
	}, nil
}

func favicon(g route.StaticGlobals) (echo.HandlerFunc, error) {
	return caching.WrapStaticHeader(echo.WrapHandler(
		http.StripPrefix("/", http.FileServer(http.FS(g.FS["Web"]))),
	), week), nil
}

func fonts(g route.StaticGlobals) (echo.HandlerFunc, error) {
	return caching.WrapStaticHeader(echo.WrapHandler(
		http.StripPrefix("/fonts/", http.FileServer(http.FS(g.FS["Fonts"]))),
	), year), nil
}

func static(g route.StaticGlobals) (echo.HandlerFunc, error) {
	return echo.WrapHandler(
		http.StripPrefix("/static/", hashfs.FileServer(g.HFS["Static"])),
	), nil
}

func app(g route.StaticGlobals) (echo.HandlerFunc, error) {
	return echo.WrapHandler(
		http.StripPrefix("/app/", hashfs.FileServer(g.HFS["App"])),
	), nil
}

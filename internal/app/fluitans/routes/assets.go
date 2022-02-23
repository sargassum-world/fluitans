package routes

import (
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/pkg/framework/httpcache"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
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
		HandlerMaker: getWebmanifest,
		Templates:    []string{"app/app.webmanifest.tmpl"},
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

func getWebmanifest(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "app/app.webmanifest.tmpl"
	err := te.RequireSegments("assets.getWebmanifest", t)
	if err != nil {
		return nil, err
	}

	return func(c echo.Context) error {
		// Render template
		c.Response().Header().Set(echo.HeaderContentType, "application/manifest+json")
		return route.Render(c, t, struct{}{}, struct{}{}, te, g)
	}, nil
}

func favicon(g route.StaticGlobals) (echo.HandlerFunc, error) {
	return echo.WrapHandler(httpcache.WrapStaticHeader(
		http.StripPrefix("/", http.FileServer(http.FS(g.FS["Web"]))),
		week)), nil
}

func fonts(g route.StaticGlobals) (echo.HandlerFunc, error) {
	return echo.WrapHandler(httpcache.WrapStaticHeader(
		http.StripPrefix("/fonts/", http.FileServer(http.FS(g.FS["Fonts"]))),
		year)), nil
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

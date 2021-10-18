package handlers

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/benbjohnson/hashfs"
	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/templates"
	"github.com/sargassum-eco/fluitans/web"
)

// Template Parameters

var embedAssets = templates.EmbedAssets{
	//nolint:gosec // This bundle is generated from code in web/app/src, so we know it's well-formed.
	BundleEagerJS:  template.JS(web.BundleEagerJS),
	BundleEagerCSS: template.CSS(web.BundleEagerCSS),
}

// Utilities

func prefix(prefix, filename string) string {
	return fmt.Sprintf("%s/%s", prefix, filename)
}

// Route Handlers

// TODO: write a template function for looking up the hashed name of any static file; include the prefix logic.
// TODO: write a template function for looking up the hashed name of any app file; include the prefix logic.
func webmanifest(c echo.Context) error {
	c.Response().Header().Set(echo.HeaderContentType, "application/manifest+json")
	return c.Render(http.StatusOK, "app.webmanifest.tmpl", struct {}{})
}

var favicon = echo.WrapHandler(
	http.FileServer(http.FS(web.StaticFS)),
)

var static = echo.WrapHandler(
	http.StripPrefix("/static/", hashfs.FileServer(web.StaticHFS)),
)

var app = echo.WrapHandler(
	http.StripPrefix("/app/", hashfs.FileServer(web.AppHFS)),
)

var fonts = echo.WrapHandler(
	http.StripPrefix("/fonts/", http.FileServer(http.FS(web.FontsFS))),
)

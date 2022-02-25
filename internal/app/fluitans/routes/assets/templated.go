// Package assets contains the route handlers for assets which are static for the server
package assets

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

func getWebmanifest(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "app/app.webmanifest.tmpl"
	te.Require(t)
	return func(c echo.Context) error {
		// Render template
		c.Response().Header().Set(echo.HeaderContentType, "application/manifest+json")
		return route.Render(c, t, struct{}{}, struct{}{}, te, g)
	}, nil
}

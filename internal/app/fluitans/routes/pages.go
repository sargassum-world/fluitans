// Package routes contains the route handlers for the Fluitans server.
package routes

import (
	"net/http"
	"os"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/networks"
	"github.com/sargassum-eco/fluitans/internal/caching"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
)

var Pages = append(
	[]route.Templated{
		{
			Path:         "/",
			Method:       http.MethodGet,
			HandlerMaker: home,
			Templates:    []string{"home.page.tmpl"},
		},
		{
			Path:         "/login",
			Method:       http.MethodGet,
			HandlerMaker: login,
			Templates:    []string{"login.page.tmpl"},
		},
	},
	networks.Pages...,
)

func home(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "home.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	return func(c echo.Context) error {
		// Handle Etag
		if noContent, err := caching.ProcessEtag(c, tte); noContent {
			return err
		}

		// Render template
		return c.Render(http.StatusOK, t, struct {
			Meta   template.Meta
			Embeds template.Embeds
		}{
			Meta: template.Meta{
				Path:       c.Request().URL.Path,
				DomainName: os.Getenv("FLUITANS_DOMAIN_NAME"),
			},
			Embeds: g.Embeds,
		})
	}, nil
}

func login(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "login.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	return func(c echo.Context) error {
		// Handle Etag
		if noContent, err := caching.ProcessEtag(c, tte); noContent {
			return err
		}

		// Render template
		return c.Render(http.StatusOK, t, struct {
			Meta   template.Meta
			Embeds template.Embeds
		}{
			Meta: template.Meta{
				Path: c.Request().URL.Path,
			},
			Embeds: g.Embeds,
		})
	}, nil
}

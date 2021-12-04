// Package routes contains the route handlers for the Fluitans server.
package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/dns"
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
			HandlerMaker: getHome,
			Templates:    []string{"home.page.tmpl"},
		},
		{
			Path:         "/login",
			Method:       http.MethodGet,
			HandlerMaker: getLogin,
			Templates:    []string{"login.page.tmpl"},
		},
	},
	append(
		networks.Pages,
		dns.Pages...,
	)...,
)

func getHome(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "home.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	switch app := g.App.(type) {
	default:
		return nil, errors.Errorf("app globals are of unexpected type %T", g.App)
	case *client.Globals:
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
					DomainName: app.Config.DomainName,
				},
				Embeds: g.Embeds,
			})
		}, nil
	}
}

func getLogin(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "login.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	switch app := g.App.(type) {
	default:
		return nil, errors.Errorf("app globals are of unexpected type %T", g.App)
	case *client.Globals:
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
					DomainName: app.Config.DomainName,
				},
				Embeds: g.Embeds,
			})
		}, nil
	}
}

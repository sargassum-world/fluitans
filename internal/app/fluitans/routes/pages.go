// Package routes contains the route handlers for the Fluitans server.
package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/dns"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/networks"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/templates"
	"github.com/sargassum-eco/fluitans/internal/route"
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

func getHome(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "home.page.tmpl"
	tte, err := templates.GetTemplate(te, t, "pages.getHome")
	if err != nil {
		return nil, err
	}

	return func(c echo.Context) error {
		// Produce output
		noContent, err := templates.ProcessEtag(c, tte, "")
		if err != nil || noContent {
			return err
		}
		return c.Render(http.StatusOK, t, templates.MakeRenderData(c, g, ""))
	}, nil
}

func getLogin(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "login.page.tmpl"
	tte, err := templates.GetTemplate(te, t, "pages.getLogin")
	if err != nil {
		return nil, err
	}

	return func(c echo.Context) error {
		// Produce output
		noContent, err := templates.ProcessEtag(c, tte, "")
		if err != nil || noContent {
			return err
		}
		return c.Render(http.StatusOK, t, templates.MakeRenderData(c, g, ""))
	}, nil
}

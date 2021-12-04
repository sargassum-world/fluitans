// Package routes contains the route handlers for the Fluitans server.
package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/dns"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/networks"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
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
	err := te.RequireSegments("pages.getHome", t)
	if err != nil {
		return nil, err
	}

	return func(c echo.Context) error {
		return route.Render(c, t, struct{}{}, te, g)
	}, nil
}

func getLogin(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "login.page.tmpl"
	err := te.RequireSegments("pages.getLogin", t)
	if err != nil {
		return nil, err
	}

	return func(c echo.Context) error {
		return route.Render(c, t, struct{}{}, te, g)
	}, nil
}

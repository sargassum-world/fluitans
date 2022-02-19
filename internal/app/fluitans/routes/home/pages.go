// Package home contains the route handlers related to the app's home screen.
package home

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

var Pages = []route.Templated{
	{
		Path:         "/",
		Method:       http.MethodGet,
		HandlerMaker: getHome,
		Templates:    []string{"home/home.page.tmpl"},
	},
}

func getHome(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "home/home.page.tmpl"
	err := te.RequireSegments("pages.getHome", t)
	if err != nil {
		return nil, err
	}

	app, ok := g.App.(*client.Globals)
	if !ok {
		return nil, client.NewUnexpectedGlobalsTypeError(g.App)
	}
	return func(c echo.Context) error {
		// Check authentication & authorization
		a, _, err := auth.GetWithSession(c, app.Clients.Sessions)
		if err != nil {
			return err
		}

		// Produce output
		return route.Render(c, t, struct{}{}, a, te, g)
	}, nil
}

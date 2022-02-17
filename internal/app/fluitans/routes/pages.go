// Package routes contains the route handlers for the Fluitans server.
package routes

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/controllers"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/dns"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/networks"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

var pageCollections = [][]route.Templated{
	{
		{
			Path:         "/",
			Method:       http.MethodGet,
			HandlerMaker: getHome,
			Templates:    []string{"home/home.page.tmpl"},
		},
	},
	AuthnPages,
	controllers.Pages,
	networks.Pages,
	dns.Pages,
}
var Pages []route.Templated = route.CollectTemplated(pageCollections)

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

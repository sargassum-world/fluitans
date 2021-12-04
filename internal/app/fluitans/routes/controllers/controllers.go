package controllers

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

func getControllers(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "controllers/controllers.page.tmpl"
	err := te.RequireSegments("controllers.getControllers", t)
	if err != nil {
		return nil, err
	}

	switch app := g.App.(type) {
	default:
		return nil, client.NewUnexpectedGlobalsTypeError(app)
	case *client.Globals:
		return func(c echo.Context) error {
			// Run queries
			controllers, err := app.Clients.ZTControllers.GetControllers()
			if err != nil {
				return err
			}

			// Produce output
			return route.Render(c, t, controllers, te, g)
		}, nil
	}
}

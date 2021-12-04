package networks

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

func getControllers(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "networks/controllers.page.tmpl"
	err := te.RequireSegments("networks.getControllers", t)
	if err != nil {
		return nil, err
	}

	switch app := g.App.(type) {
	default:
		return nil, errors.Errorf("app globals are of unexpected type %T", g.App)
	case *client.Globals:
		return func(c echo.Context) error {
			// Run queries
			controllers, err := client.GetControllers(app.Config)
			if err != nil {
				return err
			}

			// Produce output
			return route.Render(c, t, controllers, te, g)
		}, nil
	}
}

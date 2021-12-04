package networks

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/templates"
	"github.com/sargassum-eco/fluitans/internal/route"
)

func getControllers(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "networks/controllers.page.tmpl"
	tte, err := templates.GetTemplate(te, t, "networks.getControllers")
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
			noContent, err := templates.ProcessEtag(c, tte, controllers)
			if err != nil || noContent {
				return err
			}
			return c.Render(http.StatusOK, t, templates.MakeRenderData(c, g, controllers))
		}, nil
	}
}

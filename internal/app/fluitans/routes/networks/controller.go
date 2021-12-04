package networks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/conf"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/models"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

type ControllerData struct {
	Controller       models.Controller
	Status           zerotier.Status
	ControllerStatus zerotier.ControllerStatus
	Networks         map[string]zerotier.ControllerNetwork
}

func getControllerData(
	ctx context.Context, name string, templateName string,
	config conf.Config, cache *client.Cache,
) (*ControllerData, error) {
	controller, ok, err := client.FindController(name, config)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, echo.NewHTTPError(
			http.StatusNotFound,
			fmt.Sprintf("Controller %s not found for %s", name, templateName),
		)
	}

	status, controllerStatus, networkIDs, err := client.GetController(
		ctx, *controller, cache,
	)
	if err != nil {
		return nil, err
	}

	networks, err := client.GetNetworks(
		ctx, []models.Controller{*controller}, [][]string{networkIDs},
	)
	if err != nil {
		return nil, err
	}

	return &ControllerData{
		Controller:       *controller,
		Status:           *status,
		ControllerStatus: *controllerStatus,
		Networks:         networks[0],
	}, nil
}

func getController(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "networks/controller.page.tmpl"
	err := te.RequireSegments("networks.getController", t)
	if err != nil {
		return nil, err
	}

	switch app := g.App.(type) {
	default:
		return nil, errors.Errorf("app globals are of unexpected type %T", g.App)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()

			// Parse params
			name := c.Param("name")

			// Run queries
			controllerData, err := getControllerData(ctx, name, t, app.Config, app.Cache)
			if err != nil {
				return err
			}

			// Produce output
			// Zero out clocks, since they will always change the Etag
			*controllerData.Status.Clock = 0
			*controllerData.ControllerStatus.Clock = 0
			return route.Render(c, t, *controllerData, te, g)
		}, nil
	}
}

package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	ztc "github.com/sargassum-eco/fluitans/internal/clients/zerotier"
	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

type ControllerData struct {
	Controller       ztcontrollers.Controller
	Status           zerotier.Status
	ControllerStatus zerotier.ControllerStatus
	Networks         map[string]zerotier.ControllerNetwork
}

func getControllerData(
	ctx context.Context, name string, cc *ztcontrollers.Client, c *ztc.Client,
) (*ControllerData, error) {
	controller, err := cc.FindController(name)
	if err != nil {
		return nil, err
	}
	if controller == nil {
		return nil, echo.NewHTTPError(
			http.StatusNotFound, fmt.Sprintf("zerotier controller %s not found", name),
		)
	}

	status, controllerStatus, networkIDs, err := c.GetControllerInfo(ctx, *controller, cc)
	if err != nil {
		return nil, err
	}

	networks, err := c.GetAllNetworks(
		ctx, []ztcontrollers.Controller{*controller}, [][]string{networkIDs},
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
	t := "controllers/controller.page.tmpl"
	err := te.RequireSegments("controllers.getController", t)
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

		// Extract context
		ctx := c.Request().Context()

		// Parse params
		name := c.Param("name")

		// Run queries
		controllerData, err := getControllerData(
			ctx, name, app.Clients.ZTControllers, app.Clients.Zerotier,
		)
		if err != nil {
			return err
		}

		// Produce output
		// Zero out clocks, since they will always change the Etag
		*controllerData.Status.Clock = 0
		*controllerData.ControllerStatus.Clock = 0
		return route.Render(c, t, *controllerData, a, te, g)
	}, nil
}

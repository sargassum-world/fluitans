package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

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
	controller, ok, err := cc.FindController(name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, echo.NewHTTPError(
			http.StatusNotFound, fmt.Sprintf("zerotier controller %s not found", name),
		)
	}
	client, err := controller.NewClient()
	if err != nil {
		return nil, err
	}

	var status *zerotier.Status
	var controllerStatus *zerotier.ControllerStatus
	var networkIDs []string
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		res, cerr := client.GetStatusWithResponse(ctx)
		if cerr != nil {
			return err
		}
		status = res.JSON200
		return cc.Cache.SetControllerByAddress(*status.Address, *controller)
	})
	eg.Go(func() error {
		res, cerr := client.GetControllerStatusWithResponse(ctx)
		if cerr != nil {
			return err
		}
		controllerStatus = res.JSON200
		return err
	})
	eg.Go(func() error {
		res, cerr := client.GetControllerNetworksWithResponse(ctx)
		if cerr != nil {
			return err
		}
		networkIDs = *res.JSON200
		return nil
	})
	if err = eg.Wait(); err != nil {
		return nil, err
	}

	networks, err := c.GetNetworks(
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

	switch app := g.App.(type) {
	default:
		return nil, client.NewUnexpectedGlobalsTypeError(app)
	case *client.Globals:
		return func(c echo.Context) error {
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
			return route.Render(c, t, *controllerData, te, g)
		}, nil
	}
}

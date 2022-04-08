package controllers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	ztc "github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/zerotier"
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

func (h *Handlers) HandleControllerGet() auth.HTTPHandlerFunc {
	t := "controllers/controller.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		name := c.Param("name")

		// Run queries
		controllerData, err := getControllerData(c.Request().Context(), name, h.ztcc, h.ztc)
		if err != nil {
			return err
		}

		// Produce output
		// Zero out clocks before computing etag for client-side caching
		*controllerData.Status.Clock = 0
		*controllerData.ControllerStatus.Clock = 0
		return h.r.CacheablePage(c.Response(), c.Request(), t, *controllerData, a)
	}
}

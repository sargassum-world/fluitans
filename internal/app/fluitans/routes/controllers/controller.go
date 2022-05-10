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

type ControllerViewData struct {
	Controller       ztcontrollers.Controller
	Status           zerotier.Status
	ControllerStatus zerotier.ControllerStatus
	Networks         map[string]zerotier.ControllerNetwork
}

func getControllerViewData(
	ctx context.Context, name string, cc *ztcontrollers.Client, c *ztc.Client,
) (vd ControllerViewData, err error) {
	controller, err := cc.FindController(name)
	if err != nil {
		return ControllerViewData{}, err
	}
	if controller == nil {
		return ControllerViewData{}, echo.NewHTTPError(
			http.StatusNotFound, fmt.Sprintf("zerotier controller %s not found", name),
		)
	}
	vd.Controller = *controller

	status, controllerStatus, networkIDs, err := c.GetControllerInfo(ctx, *controller, cc)
	if err != nil {
		return ControllerViewData{}, err
	}
	vd.Status = *status
	vd.ControllerStatus = *controllerStatus

	networks, err := c.GetAllNetworks(
		ctx, []ztcontrollers.Controller{*controller}, [][]string{networkIDs},
	)
	if err != nil {
		return ControllerViewData{}, err
	}
	vd.Networks = networks[0]

	return vd, nil
}

func (h *Handlers) HandleControllerGet() auth.HTTPHandlerFunc {
	t := "controllers/controller.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		name := c.Param("name")

		// Run queries
		controllerViewData, err := getControllerViewData(c.Request().Context(), name, h.ztcc, h.ztc)
		if err != nil {
			return err
		}

		// Produce output
		// Zero out clocks before computing etag for client-side caching
		*controllerViewData.Status.Clock = 0
		*controllerViewData.ControllerStatus.Clock = 0
		return h.r.CacheablePage(c.Response(), c.Request(), t, controllerViewData, a)
	}
}

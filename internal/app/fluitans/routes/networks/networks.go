package networks

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

type NetworksViewData struct {
	Controller ztcontrollers.Controller
	Networks   map[string]zerotier.ControllerNetwork
}

func getNetworksViewData(
	ctx context.Context, c *ztc.Client, cc *ztcontrollers.Client,
) ([]NetworksViewData, error) {
	controllers, err := cc.GetControllers()
	if err != nil {
		return nil, err
	}

	networkIDs, err := c.GetAllNetworkIDs(ctx, controllers, cc)
	if err != nil {
		return nil, err
	}

	networks, err := c.GetAllNetworks(ctx, controllers, networkIDs)
	if err != nil {
		return nil, err
	}

	networksViewData := make([]NetworksViewData, len(controllers))
	for i, controller := range controllers {
		networksViewData[i].Controller = controller
		networksViewData[i].Networks = networks[i]
	}
	return networksViewData, nil
}

func (h *Handlers) HandleNetworksGet() auth.HTTPHandlerFunc {
	t := "networks/networks.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Run queries
		networksViewData, err := getNetworksViewData(c.Request().Context(), h.ztc, h.ztcc)
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, networksViewData, a)
	}
}

func (h *Handlers) HandleNetworksPost() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse params
		name := c.FormValue("controller")
		if name == "" {
			return echo.NewHTTPError(
				http.StatusBadRequest, "zerotier controller name not specified",
			)
		}

		// Run queries
		controller, err := h.ztcc.FindController(name)
		if err != nil {
			return err
		}
		if controller == nil {
			return echo.NewHTTPError(
				http.StatusNotFound, fmt.Sprintf("zerotier controller %s not found", name),
			)
		}

		createdNetwork, err := h.ztc.CreateNetwork(c.Request().Context(), *controller, h.ztcc)
		if err != nil {
			return err
		}
		created := createdNetwork.Id
		if created == nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "network status unknown")
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s", *created))
	}
}

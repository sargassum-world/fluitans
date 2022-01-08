package networks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	ztc "github.com/sargassum-eco/fluitans/internal/clients/zerotier"
	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

type NetworksData struct {
	Controller ztcontrollers.Controller
	Networks   map[string]zerotier.ControllerNetwork
}

func getNetworksData(
	ctx context.Context, c *ztc.Client, cc *ztcontrollers.Client,
) ([]NetworksData, error) {
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

	networksData := make([]NetworksData, len(controllers))
	for i, controller := range controllers {
		networksData[i].Controller = controller
		networksData[i].Networks = networks[i]
	}
	return networksData, nil
}

func getNetworks(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "networks/networks.page.tmpl"
	err := te.RequireSegments("networks.getNetwork", t)
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

			// Run queries
			networksData, err := getNetworksData(ctx, app.Clients.Zerotier, app.Clients.ZTControllers)
			if err != nil {
				return err
			}

			// Produce output
			return route.Render(c, t, networksData, te, g)
		}, nil
	}
}

func postNetworks(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	switch app := g.App.(type) {
	default:
		return nil, client.NewUnexpectedGlobalsTypeError(app)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()

			// Parse params
			name := c.FormValue("controller")
			if name == "" {
				return echo.NewHTTPError(
					http.StatusBadRequest, "zerotier controller name not specified",
				)
			}

			// Run queries
			controller, err := app.Clients.ZTControllers.FindController(name)
			if err != nil {
				return err
			}
			if controller == nil {
				return echo.NewHTTPError(
					http.StatusNotFound, fmt.Sprintf("zerotier controller %s not found", name),
				)
			}

			createdNetwork, err := app.Clients.Zerotier.CreateNetwork(
				ctx, *controller, app.Clients.ZTControllers,
			)
			if err != nil {
				return err
			}
			created := createdNetwork.Id
			if created == nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "network status unknown")
			}
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s", *created))
		}, nil
	}
}

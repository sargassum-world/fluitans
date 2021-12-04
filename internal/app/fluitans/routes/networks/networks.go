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
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/templates"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

type NetworksData struct {
	Controller models.Controller
	Networks   map[string]zerotier.ControllerNetwork
}

func getNetworksData(
	ctx context.Context, config conf.Config, cache *client.Cache,
) ([]NetworksData, error) {
	controllers, err := client.GetControllers(config)
	if err != nil {
		return nil, err
	}

	networkIDs, err := client.GetNetworkIDs(ctx, controllers, cache)
	if err != nil {
		return nil, err
	}

	networks, err := client.GetNetworks(ctx, controllers, networkIDs)
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

func getNetworks(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "networks/networks.page.tmpl"
	tte, err := templates.GetTemplate(te, t, "networks.getNetworks")
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

			// Run queries
			networksData, err := getNetworksData(ctx, app.Config, app.Cache)
			if err != nil {
				return err
			}

			// Produce output
			noContent, err := templates.ProcessEtag(c, tte, networksData)
			if noContent || (err != nil) {
				return err
			}
			return c.Render(http.StatusOK, t, templates.MakeRenderData(c, g, networksData))
		}, nil
	}
}

func postNetworks(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	switch app := g.App.(type) {
	default:
		return nil, errors.Errorf("app globals are of unexpected type %T", g.App)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()

			// Parse params
			name := c.FormValue("controller")
			if name == "" {
				return echo.NewHTTPError(
					http.StatusBadRequest, "Controller name not specified",
				)
			}

			// Run queries
			controller, ok, err := client.FindController(name, app.Config)
			if err != nil {
				return err
			}
			if !ok {
				return echo.NewHTTPError(
					http.StatusNotFound, fmt.Sprintf("Controller %s not found", name),
				)
			}

			createdNetwork, err := client.CreateNetwork(ctx, *controller)
			if err != nil {
				return err
			}

			created := createdNetwork.Id
			if created == nil {
				return echo.NewHTTPError(
					http.StatusInternalServerError, "Network status unknown",
				)
			}
			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s", *created))
		}, nil
	}
}

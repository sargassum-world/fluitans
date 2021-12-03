package networks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/caching"
	"github.com/sargassum-eco/fluitans/internal/fingerprint"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

type NetworksData struct {
	Controller client.Controller
	Networks   map[string]zerotier.ControllerNetwork
}

func getNetworksData(ctx context.Context, cache *client.Cache) ([]NetworksData, error) {
	controllers, err := client.GetControllers()
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
	tte, ok := te[t]
	if !ok {
		return nil, errors.Wrap(
			te.NewNotFoundError(t), "couldn't find template for networks.getNetworks",
		)
	}

	switch app := g.App.(type) {
	default:
		return nil, errors.Errorf("app globals are of unexpected type %T", g.App)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()

			// Run queries
			networksData, err := getNetworksData(ctx, app.Cache)
			if err != nil {
				return err
			}

			// Handle Etag
			etagData, err := json.Marshal(networksData)
			if err != nil {
				return err
			}

			if noContent, err := caching.ProcessEtag(c, tte, fingerprint.Compute(etagData)); noContent {
				return err
			}

			// Render template
			return c.Render(http.StatusOK, t, struct {
				Meta   template.Meta
				Embeds template.Embeds
				Data   []NetworksData
			}{
				Meta: template.Meta{
					Path:       c.Request().URL.Path,
					DomainName: client.GetEnvVarDomainName(),
				},
				Embeds: g.Embeds,
				Data:   networksData,
			})
		}, nil
	}
}

func postNetworks(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
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
		controller, ok, err := client.FindController(name)
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

		// Render template
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s", *created))
	}, nil
}

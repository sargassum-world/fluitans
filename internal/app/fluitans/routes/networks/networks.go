package networks

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-eco/fluitans/internal/caching"
	"github.com/sargassum-eco/fluitans/internal/fingerprint"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

type ControllerNetworks struct {
	Controller Controller `json:"controller"`
	Networks   []string   `json:"network"`
}

func getNetworks(c echo.Context, controllers []Controller) ([][]string, error) {
	eg, ctx := errgroup.WithContext(c.Request().Context())
	networks := make([][]string, len(controllers))
	for i, controller := range controllers {
		eg.Go(func(i int) func() error {
			return func() error {
				client, cerr := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
				if cerr != nil {
					return nil
				}

				res, err := client.GetControllerNetworksWithResponse(ctx)
				if err != nil {
					return err
				}

				networks[i] = *res.JSON200
				return nil
			}
		}(i))
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return networks, nil
}

func networks(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "networks.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	return func(c echo.Context) error {
		// Run queries
		networks, err := getNetworks(c, storedControllers)
		if err != nil {
			return err
		}

		controllerNetworks := make([]ControllerNetworks, len(storedControllers))
		for i, controller := range storedControllers {
			controllerNetworks[i].Controller = controller
			controllerNetworks[i].Networks = networks[i]
		}

		// TODO: look up network names, too

		// Handle Etag
		etagData, err := json.Marshal(controllerNetworks)
		if err != nil {
			return err
		}

		if noContent, err := caching.ProcessEtag(c, tte, fingerprint.Compute(etagData)); noContent {
			return err
		}

		// Render template
		return c.Render(http.StatusOK, t, struct {
			Meta               template.Meta
			Embeds             template.Embeds
			ControllerNetworks []ControllerNetworks
		}{
			Meta: template.Meta{
				Path: c.Request().URL.Path,
			},
			Embeds:             g.Embeds,
			ControllerNetworks: controllerNetworks,
		})
	}, nil
}

func createNetwork(c echo.Context, controller Controller) (*zerotier.ControllerNetwork, error) {
	client, cerr := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
	if cerr != nil {
		return nil, cerr
	}

	ctx := c.Request().Context()
	sRes, err := client.GetStatusWithResponse(ctx)
	if err != nil {
		return nil, err
	}
	status := *sRes.JSON200

	private := true
	n6plane := true
	v6AssignMode := zerotier.V6AssignMode{
		N6plane: &n6plane,
		Rfc4193: nil,
		Zt:      nil,
	}

	body := zerotier.GenerateControllerNetworkJSONRequestBody{}
	body.Private = &private
	body.V6AssignMode = &v6AssignMode

	nRes, err := client.GenerateControllerNetworkWithResponse(
		ctx, *status.Address, body,
	)
	if err != nil {
		return nil, err
	}

	return nRes.JSON200, nil
}

func postNetworks(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	return func(c echo.Context) error {
		// Parse params
		name := c.FormValue("controller")
		if name == "" {
			return echo.NewHTTPError(
				http.StatusBadRequest, "Controller name not specified",
			)
		}

		// Run queries
		controller, ok := findController(storedControllers, name)
		if !ok {
			return echo.NewHTTPError(
				http.StatusNotFound, fmt.Sprintf("Controller %s not found", name),
			)
		}

		createdNetwork, err := createNetwork(c, *controller)
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

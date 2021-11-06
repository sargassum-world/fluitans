package networks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-eco/fluitans/internal/caching"
	"github.com/sargassum-eco/fluitans/internal/fingerprint"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

func getNetworkIDs(c echo.Context, controllers []Controller) ([][]string, error) {
	eg, ctx := errgroup.WithContext(c.Request().Context())
	networkIDs := make([][]string, len(controllers))
	for i := range controllers {
		networkIDs[i] = []string{}
	}
	for i, controller := range controllers {
		eg.Go(func(i int, controller Controller) func() error {
			return func() error {
				client, cerr := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
				if cerr != nil {
					return nil
				}

				res, err := client.GetControllerNetworksWithResponse(ctx)
				if err != nil {
					return err
				}

				networkIDs[i] = *res.JSON200
				return nil
			}
		}(i, controller))
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return networkIDs, nil
}

func getNetworks(
	c echo.Context, controllers []Controller, networkIDs [][]string,
) ([]map[string]zerotier.ControllerNetwork, error) {
	eg, ctx := errgroup.WithContext(c.Request().Context())
	networks := make([][]zerotier.ControllerNetwork, len(controllers))
	for i := range controllers {
		networks[i] = make([]zerotier.ControllerNetwork, len(networkIDs[i]))
		for j := range networkIDs[i] {
			networks[i][j] = zerotier.ControllerNetwork{}
		}
	}
	for i, controller := range controllers {
		client, cerr := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
		for j, id := range networkIDs[i] {
			eg.Go(func(i int, client *zerotier.ClientWithResponses, j int, id string) func() error {
				return func() error {
					if cerr != nil {
						return nil
					}

					res, err := client.GetControllerNetworkWithResponse(ctx, id)
					if err != nil {
						return err
					}

					networks[i][j] = *res.JSON200
					return nil
				}
			}(i, client, j, id))
		}
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	keyedNetworks := make([]map[string]zerotier.ControllerNetwork, len(controllers))
	for i := range controllers {
		keyedNetworks[i] = make(map[string]zerotier.ControllerNetwork, len(networkIDs[i]))
		for j, id := range networkIDs[i] {
			keyedNetworks[i][id] = networks[i][j]
		}
	}

	return keyedNetworks, nil
}

func networks(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "networks/networks.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	return func(c echo.Context) error {
		// Run queries
		networkIDs, err := getNetworkIDs(c, storedControllers)
		if err != nil {
			return err
		}

		networks, err := getNetworks(c, storedControllers, networkIDs)
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
			Meta   template.Meta
			Embeds template.Embeds
			Data   []ControllerNetworks
		}{
			Meta: template.Meta{
				Path:       c.Request().URL.Path,
				DomainName: os.Getenv("FLUITANS_DOMAIN_NAME"),
			},
			Embeds: g.Embeds,
			Data:   controllerNetworks,
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
	ipv4Type := 2048
	ipv4ARPType := 2054
	ipv6Type := 34525
	rules := []map[string]interface{}{
		{
			"type":      "MATCH_ETHERTYPE",
			"etherType": ipv4Type,
			"not":       true,
		},
		{
			"type":      "MATCH_ETHERTYPE",
			"etherType": ipv4ARPType,
			"not":       true,
		},
		{
			"type":      "MATCH_ETHERTYPE",
			"etherType": ipv6Type,
			"not":       true,
		},
		{
			"type": "ACTION_DROP",
		},
		{
			"type": "ACTION_ACCEPT",
		},
	}
	fmt.Println(rules)

	body := zerotier.GenerateControllerNetworkJSONRequestBody{}
	body.Private = &private
	body.V6AssignMode = &v6AssignMode
	body.Rules = &rules

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

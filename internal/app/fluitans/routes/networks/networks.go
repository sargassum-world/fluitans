package networks

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/caching"
	"github.com/sargassum-eco/fluitans/internal/fingerprint"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

type ControllerNetworks struct {
	Controller client.Controller
	Networks   map[string]zerotier.ControllerNetwork
}

func getNetworks(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "networks/networks.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	return func(c echo.Context) error {
		// Run queries
		controllers, err := client.GetControllers()
		if err != nil {
			return err
		}

		networkIDs, err := client.GetNetworkIDs(c, controllers)
		if err != nil {
			return err
		}

		networks, err := client.GetNetworks(c, controllers, networkIDs)
		if err != nil {
			return err
		}

		controllerNetworks := make([]ControllerNetworks, len(controllers))
		for i, controller := range controllers {
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
				DomainName: client.GetDomainName(),
			},
			Embeds: g.Embeds,
			Data:   controllerNetworks,
		})
	}, nil
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
		controller, ok, err := client.FindController(name)
		if err != nil {
			return err
		}

		if !ok {
			return echo.NewHTTPError(
				http.StatusNotFound, fmt.Sprintf("Controller %s not found", name),
			)
		}

		createdNetwork, err := client.CreateNetwork(c, *controller)
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

package networks

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/caching"
	"github.com/sargassum-eco/fluitans/internal/fingerprint"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

func getNetworkInfo(
	c echo.Context, controller Controller, id string,
) (*zerotier.ControllerNetwork, error) {
	client, cerr := zerotier.NewAuthClientWithResponses(
		controller.Server, controller.Authtoken,
	)
	if cerr != nil {
		return nil, cerr
	}

	res, err := client.GetControllerNetworkWithResponse(c.Request().Context(), id)
	if err != nil {
		return nil, err
	}

	return res.JSON200, nil
}

func getControllerAddress(networkID string) string {
	addressLength := 10
	return networkID[:addressLength]
}

func network(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "network.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	return func(c echo.Context) error {
		// Parse params
		id := c.Param("id")
		address := getControllerAddress(id)

		// Run queries
		controller, err := findControllerByAddress(c, storedControllers, address)
		if err != nil {
			return err
		}

		network, err := getNetworkInfo(c, *controller, id)
		if err != nil {
			return err
		}

		// Handle Etag
		etagData, err := json.Marshal(network)
		if err != nil {
			return err
		}

		if noContent, err := caching.ProcessEtag(
			c, tte, fingerprint.Compute(etagData),
		); noContent {
			return err
		}

		// Render template
		return c.Render(http.StatusOK, t, struct {
			Meta       template.Meta
			Embeds     template.Embeds
			Controller Controller
			Network    zerotier.ControllerNetwork
		}{
			Meta: template.Meta{
				Path: c.Request().URL.Path,
			},
			Embeds:     g.Embeds,
			Controller: *controller,
			Network:    *network,
		})
	}, nil
}

func deleteNetwork(c echo.Context, controller Controller, id string) error {
	client, err := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	_, err = client.DeleteControllerNetworkWithResponse(ctx, id)
	return err
}

func postNetwork(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	return func(c echo.Context) error {
		// Parse params
		id := c.Param("id")
		address := getControllerAddress(id)
		method := c.FormValue("method")

		// Run queries
		controller, err := findControllerByAddress(c, storedControllers, address)
		if err != nil {
			return err
		}

		fmt.Println(method)
		if method == "DELETE" {
			err = deleteNetwork(c, *controller, id)
			if err != nil {
				return err
			}

			// Render template
			return c.Redirect(http.StatusSeeOther, "/networks")
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s", id))
	}, nil
}

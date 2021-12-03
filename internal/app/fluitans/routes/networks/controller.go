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

type ControllerData struct {
	Controller       client.Controller
	Status           zerotier.Status
	ControllerStatus zerotier.ControllerStatus
	Networks         map[string]zerotier.ControllerNetwork
}

func getControllerData(
	ctx context.Context, name string, templateName string, cache *client.Cache,
) (*ControllerData, error) {
	controller, ok, err := client.FindController(name)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, echo.NewHTTPError(
			http.StatusNotFound,
			fmt.Sprintf("Controller %s not found for %s", name, templateName),
		)
	}

	status, controllerStatus, networkIDs, err := client.GetController(
		ctx, *controller, cache,
	)
	if err != nil {
		return nil, err
	}

	networks, err := client.GetNetworks(
		ctx, []client.Controller{*controller}, [][]string{networkIDs},
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

func getController(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "networks/controller.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, errors.Wrap(
			te.NewNotFoundError(t), "couldn't find template for networks.getController",
		)
	}

	switch app := g.App.(type) {
	default:
		return nil, errors.Errorf("app globals are of unexpected type %T", g.App)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()

			// Parse params
			name := c.Param("name")

			// Run queries
			controllerData, err := getControllerData(ctx, name, t, app.Cache)
			if err != nil {
				return err
			}

			// Handle Etag
			// Zero out clocks, since they will always change the Etag
			*controllerData.Status.Clock = 0
			*controllerData.ControllerStatus.Clock = 0
			etagData, err := json.Marshal(controllerData)
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
				Data   ControllerData
			}{
				Meta: template.Meta{
					Path:       c.Request().URL.Path,
					DomainName: client.GetEnvVarDomainName(),
				},
				Embeds: g.Embeds,
				Data:   *controllerData,
			})
		}, nil
	}
}

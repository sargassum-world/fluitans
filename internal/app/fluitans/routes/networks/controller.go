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

type ControllerNetworks struct {
	Controller Controller                            `json:"controller"`
	Networks   map[string]zerotier.ControllerNetwork `json:"network"`
}

func getControllerInfo(
	c echo.Context, controller Controller,
) (*zerotier.Status, *zerotier.ControllerStatus, []string, error) {
	client, cerr := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
	if cerr != nil {
		return nil, nil, nil, cerr
	}

	var status *zerotier.Status
	var controllerStatus *zerotier.ControllerStatus
	var networks []string
	eg, ctx := errgroup.WithContext(c.Request().Context())
	eg.Go(func() error {
		res, err := client.GetStatusWithResponse(ctx)
		if err != nil {
			return err
		}

		status = res.JSON200
		return nil
	})
	eg.Go(func() error {
		res, err := client.GetControllerStatusWithResponse(ctx)
		if err != nil {
			return err
		}

		controllerStatus = res.JSON200
		return err
	})
	eg.Go(func() error {
		res, err := client.GetControllerNetworksWithResponse(ctx)
		if err != nil {
			return err
		}

		networks = *res.JSON200
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, nil, nil, err
	}

	return status, controllerStatus, networks, nil
}

type ControllerData struct {
	Controller       Controller
	Status           zerotier.Status
	ControllerStatus zerotier.ControllerStatus
	Networks         map[string]zerotier.ControllerNetwork
}

func getControllerData(c echo.Context, name string, templateName string) (*ControllerData, error) {
	controller, ok := findController(storedControllers, name)
	if !ok {
		return nil, echo.NewHTTPError(
			http.StatusNotFound,
			fmt.Sprintf("Controller %s not found for %s", name, templateName),
		)
	}

	status, controllerStatus, networkIDs, err := getControllerInfo(c, *controller)
	if err != nil {
		return nil, err
	}

	networks, err := getNetworks(c, []Controller{*controller}, [][]string{networkIDs})
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

func controller(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "networks/controller.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	return func(c echo.Context) error {
		// Parse params
		name := c.Param("name")

		// Run queries
		controllerData, err := getControllerData(c, name, t)
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
				DomainName: os.Getenv("FLUITANS_DOMAIN_NAME"),
			},
			Embeds: g.Embeds,
			Data:   *controllerData,
		})
	}, nil
}

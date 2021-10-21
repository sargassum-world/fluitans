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

func makeControllerDataEtagSegment(
	controller Controller,
	status zerotier.Status,
	controllerStatus zerotier.ControllerStatus,
	networks []string,
) (string, error) {
	// Zero out clocks, since they will always change the Etag
	*status.Clock = 0
	*controllerStatus.Clock = 0

	// Combine all variable data for the Etag
	etagData, err := json.Marshal(struct {
		Controller       Controller                `json:"controller"`
		Status           zerotier.Status           `json:"status"`
		ControllerStatus zerotier.ControllerStatus `json:"controllerStatus"`
		Networks         []string                  `json:"network"`
	}{
		Controller:       controller,
		Status:           status,
		ControllerStatus: controllerStatus,
		Networks:         networks,
	})
	if err != nil {
		return "", err
	}

	return fingerprint.Compute(etagData), nil
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

func controller(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "controller.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	return func(c echo.Context) error {
		// Parse params
		name := c.Param("name")

		// Run queries
		controller, ok := findController(storedControllers, name)
		if !ok {
			return echo.NewHTTPError(
				http.StatusNotFound,
				fmt.Sprintf("Controller %s not found for %s", name, t),
			)
		}

		status, controllerStatus, networks, err := getControllerInfo(c, *controller)
		if err != nil {
			return err
		}

		// Handle Etag
		dataEtagSegment, err := makeControllerDataEtagSegment(
			*controller, *status, *controllerStatus, networks,
		)
		if err != nil {
			return err
		}

		if noContent, err := caching.ProcessEtag(c, tte, dataEtagSegment); noContent {
			return err
		}

		// Render template
		return c.Render(http.StatusOK, t, struct {
			Meta             template.Meta
			Embeds           template.Embeds
			Controller       Controller
			Status           zerotier.Status
			ControllerStatus zerotier.ControllerStatus
			Networks         []string
		}{
			Meta: template.Meta{
				Path: c.Request().URL.Path,
			},
			Embeds:           g.Embeds,
			Controller:       *controller,
			Status:           *status,
			ControllerStatus: *controllerStatus,
			Networks:         networks,
		})
	}, nil
}

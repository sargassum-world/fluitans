package networks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/models"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

func setMemberAuthorization(
	ctx context.Context, controller models.Controller, networkID string,
	memberAddress string, authorized bool,
) error {
	auth := authorized
	err := client.UpdateMember(
		ctx, controller, networkID, memberAddress,
		zerotier.SetControllerNetworkMemberJSONRequestBody{Authorized: &auth},
	)
	return err
}

func postDevice(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	switch app := g.App.(type) {
	default:
		return nil, errors.Errorf("app globals are of unexpected type %T", g.App)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()
			l := c.Logger()

			// Parse params
			networkID := c.Param("id")
			controllerAddress := client.GetControllerAddress(networkID)
			memberAddress := c.Param("address")
			method := c.FormValue("method")

			// Run queries
			controller, err := client.FindControllerByAddress(
				ctx, controllerAddress, app.Config, app.Cache, l,
			)
			if err != nil {
				return err
			}

			switch method {
			default:
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"invalid POST method %s", method,
				))
			case "AUTHORIZE":

				if err = setMemberAuthorization(
					ctx, *controller, networkID, memberAddress, true,
				); err != nil {
					return err
				}
			case "DEAUTHORIZE":
				if err = setMemberAuthorization(
					ctx, *controller, networkID, memberAddress, false,
				); err != nil {
					return err
				}
			}

			return c.Redirect(http.StatusSeeOther, fmt.Sprintf(
				"/networks/%s#device-%s", networkID, memberAddress,
			))
		}, nil
	}
}

func postDevices(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	switch app := g.App.(type) {
	default:
		return nil, errors.Errorf("app globals are of unexpected type %T", g.App)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()
			l := c.Logger()

			// Parse params
			networkID := c.Param("id")
			controllerAddress := client.GetControllerAddress(networkID)
			memberAddress := c.FormValue("address")

			// Run queries
			controller, err := client.FindControllerByAddress(
				ctx, controllerAddress, app.Config, app.Cache, l,
			)
			if err != nil {
				return err
			}

			err = setMemberAuthorization(ctx, *controller, networkID, memberAddress, true)
			if err != nil {
				return err
			}

			return c.Redirect(
				http.StatusSeeOther,
				fmt.Sprintf("/networks/%s#device-%s", networkID, memberAddress),
			)
		}, nil
	}
}

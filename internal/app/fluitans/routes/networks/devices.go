package networks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	ztc "github.com/sargassum-eco/fluitans/internal/clients/zerotier"
	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

func setMemberAuthorization(
	ctx context.Context, controller ztcontrollers.Controller, networkID string,
	memberAddress string, authorized bool, c *ztc.Client,
) error {
	auth := authorized
	if err := c.UpdateMember(
		ctx, controller, networkID, memberAddress,
		zerotier.SetControllerNetworkMemberJSONRequestBody{Authorized: &auth},
	); err != nil {
		return err
	}
	if authorized {
		// We might've added a new network ID, so we should invalidate the cache
		c.Cache.UnsetNetworkMembersByID(networkID)
	}
	return nil
}

func postDevice(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	switch app := g.App.(type) {
	default:
		return nil, client.NewUnexpectedGlobalsTypeError(app)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()

			// Parse params
			networkID := c.Param("id")
			controllerAddress := ztc.GetControllerAddress(networkID)
			memberAddress := c.Param("address")
			method := c.FormValue("method")

			// Run queries
			controller, err := app.Clients.ZTControllers.FindControllerByAddress(ctx, controllerAddress)
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
					ctx, *controller, networkID, memberAddress, true, app.Clients.Zerotier,
				); err != nil {
					return err
				}
			case "DEAUTHORIZE":
				if err = setMemberAuthorization(
					ctx, *controller, networkID, memberAddress, false, app.Clients.Zerotier,
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

func postDevices(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	switch app := g.App.(type) {
	default:
		return nil, client.NewUnexpectedGlobalsTypeError(app)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()

			// Parse params
			networkID := c.Param("id")
			controllerAddress := ztc.GetControllerAddress(networkID)
			memberAddress := c.FormValue("address")

			// Run queries
			controller, err := app.Clients.ZTControllers.FindControllerByAddress(ctx, controllerAddress)
			if err != nil {
				return err
			}
			if err = setMemberAuthorization(
				ctx, *controller, networkID, memberAddress, true, app.Clients.Zerotier,
			); err != nil {
				return err
			}
			return c.Redirect(
				http.StatusSeeOther, fmt.Sprintf("/networks/%s#device-%s", networkID, memberAddress),
			)
		}, nil
	}
}

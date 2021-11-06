package networks

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

func setMemberAuthorization(
	c echo.Context, controller Controller, networkID string, memberAddress string, authorized bool,
) error {
	client, err := zerotier.NewAuthClientWithResponses(controller.Server, controller.Authtoken)
	if err != nil {
		return err
	}

	ctx := c.Request().Context()
	auth := authorized
	_, err = client.SetControllerNetworkMemberWithResponse(
		ctx, networkID, memberAddress,
		zerotier.SetControllerNetworkMemberJSONRequestBody{Authorized: &auth},
	)
	if err != nil {
		return err
	}
	return nil
}

func postDevice(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	return func(c echo.Context) error {
		// Parse params
		networkID := c.Param("id")
		controllerAddress := getControllerAddress(networkID)
		memberAddress := c.Param("address")
		method := c.FormValue("method")

		// Run queries
		controller, err := findControllerByAddress(c, storedControllers, controllerAddress)
		if err != nil {
			return err
		}

		switch method {
		case "AUTHORIZE":
			err = setMemberAuthorization(c, *controller, networkID, memberAddress, true)
			if err != nil {
				return err
			}
		case "DEAUTHORIZE":
			err = setMemberAuthorization(c, *controller, networkID, memberAddress, false)
			if err != nil {
				return err
			}
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s#device-%s", networkID, memberAddress))
	}, nil
}

func postDevices(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	return func(c echo.Context) error {
		// Parse params
		networkID := c.Param("id")
		controllerAddress := getControllerAddress(networkID)
		memberAddress := c.FormValue("address")

		// Run queries
		controller, err := findControllerByAddress(c, storedControllers, controllerAddress)
		if err != nil {
			return err
		}

		err = setMemberAuthorization(c, *controller, networkID, memberAddress, true)
		if err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s#device-%s", networkID, memberAddress))
	}, nil
}

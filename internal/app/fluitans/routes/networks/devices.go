package networks

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

func setMemberAuthorization(
	c echo.Context,
	controller client.Controller,
	networkID string,
	memberAddress string,
	authorized bool,
) error {
	auth := authorized
	err := client.UpdateMember(
		c, controller, networkID, memberAddress,
		zerotier.SetControllerNetworkMemberJSONRequestBody{Authorized: &auth},
	)
	return err
}

func postDevice(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	switch cache := g.Cache.(type) {
	default:
		return nil, fmt.Errorf("global cache is of unexpected type %T", g.Cache)
	case *client.Cache:
		return func(c echo.Context) error {
			// Parse params
			networkID := c.Param("id")
			controllerAddress := client.GetControllerAddress(networkID)
			memberAddress := c.Param("address")
			method := c.FormValue("method")

			// Run queries
			controller, err := client.FindControllerByAddress(c, controllerAddress, cache)
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
}

func postDevices(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	switch cache := g.Cache.(type) {
	default:
		return nil, fmt.Errorf("global cache is of unexpected type %T", g.Cache)
	case *client.Cache:
		return func(c echo.Context) error {
			// Parse params
			networkID := c.Param("id")
			controllerAddress := client.GetControllerAddress(networkID)
			memberAddress := c.FormValue("address")

			// Run queries
			controller, err := client.FindControllerByAddress(c, controllerAddress, cache)
			if err != nil {
				return err
			}

			err = setMemberAuthorization(c, *controller, networkID, memberAddress, true)
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

package networks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/clients/desec"
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

func setMemberName(
	ctx context.Context, controller ztcontrollers.Controller, networkID string,
	memberAddress, memberName string, c *ztc.Client, dc *desec.Client,
) error {
	network, memberAddresses, err := c.GetNetworkInfo(ctx, controller, networkID)
	if err != nil {
		return err
	}

	// Ensure we're allowed to set a name
	networkName := *network.Name
	named, err := checkNamedByDNS(ctx, networkName, networkID, dc)
	if err != nil {
		return err
	}
	if !named {
		return fmt.Errorf("Network does not have a valid domain name for naming members.")
	}
	hasMember := false
	for _, address := range memberAddresses {
		if address == memberAddress {
			hasMember = true
			break
		}
	}
	if !hasMember {
		return fmt.Errorf("Cannot set domain name for device which is not a network member.")
	}

	n6PlaneAddress, err := zerotier.Get6Plane(networkID, memberAddress)
	if err != nil {
		return err
	}
	fqdn := fmt.Sprintf("%s.d.%s", memberName, networkName)
	subname := strings.TrimSuffix(fqdn, fmt.Sprintf(".%s", dc.Config.DomainName))
	if _, err := dc.CreateRRset(
		ctx, subname, "AAAA", c.Config.DNS.DeviceTTL, []string{n6PlaneAddress},
	); err != nil {
		return err
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
			case "SETNAME":
				memberName := c.FormValue("name")
				if err = setMemberName(
					ctx, *controller, networkID, memberAddress, memberName,
					app.Clients.Zerotier, app.Clients.Desec,
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
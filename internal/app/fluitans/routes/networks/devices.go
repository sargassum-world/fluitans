package networks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/internal/clients/desec"
	ztc "github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/zerotier"
)

// Authorization

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

func (h *Handlers) HandleDeviceAuthorizationPost() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse params
		networkID := c.Param("id")
		controllerAddress := ztc.GetControllerAddress(networkID)
		memberAddress := c.Param("address")
		authorization := strings.ToLower(c.FormValue("authorization")) == "true"

		// Run queries
		ctx := c.Request().Context()
		controller, err := h.ztcc.FindControllerByAddress(ctx, controllerAddress)
		if err != nil {
			return err
		}
		if err = setMemberAuthorization(
			ctx, *controller, networkID, memberAddress, authorization, h.ztc,
		); err != nil {
			return err
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf(
			"/networks/%s#device-%s", networkID, memberAddress,
		))
	}
}

// Naming

func confirmMemberNameManageable(
	ctx context.Context, controller ztcontrollers.Controller, networkID string,
	memberAddress, memberName string, c *ztc.Client, dc *desec.Client,
) (memberSubname string, err error) {
	network, memberAddresses, err := c.GetNetworkInfo(ctx, controller, networkID)
	if err != nil {
		return "", err
	}

	networkName := *network.Name
	named, err := checkNamedByDNS(ctx, networkName, networkID, dc)
	if err != nil {
		return "", err
	}
	if !named {
		return "", errors.Errorf("network does not have a valid domain name for naming members")
	}
	hasMember := false
	for _, address := range memberAddresses {
		if address == memberAddress {
			hasMember = true
			break
		}
	}
	if !hasMember {
		return "", errors.Errorf(
			"cannot set domain name for device which is not a network member",
		)
	}

	fqdn := fmt.Sprintf("%s.d.%s", memberName, networkName)
	return strings.TrimSuffix(fqdn, fmt.Sprintf(".%s", dc.Config.DomainName)), nil
}

func setMemberName(
	ctx context.Context, controller ztcontrollers.Controller, networkID string,
	memberAddress, memberName string, c *ztc.Client, dc *desec.Client,
) error {
	memberSubname, err := confirmMemberNameManageable(
		ctx, controller, networkID, memberAddress, memberName, c, dc,
	)
	if err != nil {
		return err
	}

	n6PlaneAddress, err := zerotier.Get6Plane(networkID, memberAddress)
	if err != nil {
		return err
	}
	// TODO: prohibit assigning a name which was already assigned
	if _, err := dc.CreateRRset(
		ctx, memberSubname, "AAAA", c.Config.DNS.DeviceTTL, []string{n6PlaneAddress},
	); err != nil {
		return err
	}
	return nil
}

func unsetMemberName(
	ctx context.Context, controller ztcontrollers.Controller, networkID string,
	memberAddress, memberName string, c *ztc.Client, dc *desec.Client,
) error {
	memberSubname, err := confirmMemberNameManageable(
		ctx, controller, networkID, memberAddress, memberName, c, dc,
	)
	if err != nil {
		return err
	}

	// TODO: first confirm that the RRset contains an IP address associated with the member
	if err := dc.DeleteRRset(ctx, memberSubname, "AAAA"); err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleDeviceNamePost() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse params
		networkID := c.Param("id")
		controllerAddress := ztc.GetControllerAddress(networkID)
		memberAddress := c.Param("address")
		setName := c.FormValue("set-name")

		// Run queries
		ctx := c.Request().Context()
		controller, err := h.ztcc.FindControllerByAddress(ctx, controllerAddress)
		if err != nil {
			return err
		}

		switch setName {
		default:
			if err = setMemberName(
				ctx, *controller, networkID, memberAddress, setName, h.ztc, h.dc,
			); err != nil {
				return err
			}
		case "":
			nameToUnset := c.FormValue("unset-name")
			if err = unsetMemberName(
				ctx, *controller, networkID, memberAddress, nameToUnset, h.ztc, h.dc,
			); err != nil {
				return err
			}
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf(
			"/networks/%s#device-%s", networkID, memberAddress,
		))
	}
}

func (h *Handlers) HandleDevicesPost() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse params
		networkID := c.Param("id")
		controllerAddress := ztc.GetControllerAddress(networkID)
		memberAddress := c.FormValue("address")

		// Run queries
		ctx := c.Request().Context()
		controller, err := h.ztcc.FindControllerByAddress(ctx, controllerAddress)
		if err != nil {
			return err
		}
		if err = setMemberAuthorization(
			ctx, *controller, networkID, memberAddress, true, h.ztc,
		); err != nil {
			return err
		}

		// Redirect user
		return c.Redirect(
			http.StatusSeeOther, fmt.Sprintf("/networks/%s#device-%s", networkID, memberAddress),
		)
	}
}

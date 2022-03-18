package networks

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/internal/clients/desec"
	ztc "github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/godest/turbo"
	"github.com/sargassum-world/fluitans/pkg/zerotier"
)

// Authorization

const devicesListPartial = "networks/devices-list.partial.tmpl"

func replaceDevicesListStream(
	ctx context.Context, controllerAddress, networkID string, a auth.Auth,
	c *ztc.Client, cc *ztcontrollers.Client, dc *desec.Client,
) (turbo.Stream, error) {
	networkData, err := getNetworkData(ctx, controllerAddress, networkID, c, cc, dc)
	if err != nil {
		return turbo.Stream{}, err
	}
	return turbo.Stream{
		Action:   turbo.StreamReplace,
		Target:   "network-" + networkID + "-devices",
		Template: devicesListPartial,
		Data: map[string]interface{}{
			"Members":    networkData.Members,
			"Network":    networkData.Network,
			"NetworkDNS": networkData.NetworkDNS,
			"Auth":       a,
		},
	}, nil
}

func (h *Handlers) HandleDevicesPost() auth.Handler {
	t := devicesListPartial
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
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

		// Render Turbo Stream if accepted
		if turbo.StreamAccepted(c.Request().Header) {
			// We send the entire devices list because the content of the devices list partial depends on
			// whether there's at least one device in the network, and this is the simplest solution which
			// handles all edge cases.
			replaceStream, err := replaceDevicesListStream(
				c.Request().Context(), controllerAddress, networkID, a, h.ztc, h.ztcc, h.dc,
			)
			if err != nil {
				return err
			}
			return h.r.TurboStreams(c.Response(), replaceStream)
		}

		// Redirect user
		return c.Redirect(
			http.StatusSeeOther, fmt.Sprintf("/networks/%s#device-%s", networkID, memberAddress),
		)
	}
}

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

func (h *Handlers) HandleDeviceAuthorizationPost() auth.Handler {
	t := devicesListPartial
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
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

		// Render Turbo Stream if accepted
		if turbo.StreamAccepted(c.Request().Header) {
			// We send the entire devices list because we already have to look up roughly the same
			// amount of data to give the device partial, and it's probably not worth the additional code
			// complexity to try to only look up the data for this device in order to send a smaller
			// HTTP response payload.
			replaceStream, err := replaceDevicesListStream(
				c.Request().Context(), controllerAddress, networkID, a, h.ztc, h.ztcc, h.dc,
			)
			if err != nil {
				return err
			}
			return h.r.TurboStreams(c.Response(), replaceStream)
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

func (h *Handlers) HandleDeviceNamePost() auth.Handler {
	return func(c echo.Context, a auth.Auth) error {
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

		// Render Turbo Stream if accepted
		if turbo.StreamAccepted(c.Request().Header) {
			// We send the entire devices list because we already have to look up roughly the same
			// amount of data to give the device partial, and it's probably not worth the additional code
			// complexity to try to only look up the data for this device in order to send a smaller
			// HTTP response payload.
			replaceStream, err := replaceDevicesListStream(
				c.Request().Context(), controllerAddress, networkID, a, h.ztc, h.ztcc, h.dc,
			)
			if err != nil {
				return err
			}
			return h.r.TurboStreams(c.Response(), replaceStream)
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf(
			"/networks/%s#device-%s", networkID, memberAddress,
		))
	}
}

package networks

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/handling"
	"github.com/sargassum-world/godest/turbostreams"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	desecc "github.com/sargassum-world/fluitans/internal/clients/desec"
	ztc "github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/desec"
	"github.com/sargassum-world/fluitans/pkg/zerotier"
)

// Devices

const devicesListPartial = "networks/devices-list.partial.tmpl"

func replaceDevicesListStream(
	ctx context.Context, controllerAddress, networkID string, a auth.Auth,
	c *ztc.Client, cc *ztcontrollers.Client, dc *desecc.Client,
) (turbostreams.Message, error) {
	networkViewData, err := getNetworkViewData(ctx, controllerAddress, networkID, c, cc, dc)
	if err != nil {
		return turbostreams.Message{}, errors.Wrapf(err, "couldn't get network %s data", networkID)
	}
	return turbostreams.Message{
		Action:   turbostreams.ActionReplace,
		Target:   "/networks/" + networkID + "/devices",
		Template: devicesListPartial,
		Data: map[string]interface{}{
			"Members":    networkViewData.Members,
			"Network":    networkViewData.Network,
			"NetworkDNS": networkViewData.NetworkDNS,
			"Auth":       a,
		},
	}, nil
}

func (h *Handlers) HandleDevicesSub() turbostreams.HandlerFunc {
	return func(c *turbostreams.Context) error {
		// Parse params
		networkID := c.Param("id")
		controllerAddress := ztc.GetControllerAddress(networkID)

		// Run queries
		ctx := c.Context()
		controller, err := h.ztcc.FindControllerByAddress(ctx, controllerAddress)
		if err != nil {
			return errors.Wrapf(err, "couldn't find controller %s", controllerAddress)
		}
		network, err := h.ztc.GetNetwork(ctx, *controller, networkID)
		if err != nil {
			return errors.Wrapf(err, "couldn't get network %s", networkID)
		}
		if network == nil {
			return errors.Errorf("network %s not found", networkID)
		}

		// Allow subscription
		return nil
	}
}

func checkDevicesList(
	ctx context.Context, controllerAddress, networkID string, prevDevices StringSet,
	c *ztc.Client, cc *ztcontrollers.Client,
) (changed bool, updatedDevices StringSet, err error) {
	controller, err := cc.FindControllerByAddress(ctx, controllerAddress)
	if err != nil {
		return false, prevDevices, errors.Wrapf(err, "couldn't find controller %s", controllerAddress)
	}
	addresses, err := c.GetNetworkMemberAddresses(ctx, *controller, networkID)
	if err != nil {
		return false, prevDevices, errors.Wrapf(
			err, "couldn't get network %s member addresses", networkID,
		)
	}
	updatedDevices = NewStringSet(addresses)
	if updatedDevices.Equals(prevDevices) {
		return false, prevDevices, nil
	}
	return true, updatedDevices, nil
}

func (h *Handlers) HandleDevicesPub() turbostreams.HandlerFunc {
	t := devicesListPartial
	h.r.MustHave(t)
	return func(c *turbostreams.Context) error {
		// Make change trackers
		initialized := false
		var devices StringSet

		// Parse params
		networkID := c.Param("id")
		controllerAddress := ztc.GetControllerAddress(networkID)

		// Publish periodically
		const pubInterval = 5 * time.Second
		return handling.RepeatImmediate(c.Context(), pubInterval, func() (done bool, err error) {
			// Check for changes
			changed, updatedDevices, err := checkDevicesList(
				c.Context(), controllerAddress, networkID, devices, h.ztc, h.ztcc,
			)
			if err != nil {
				return false, errors.Wrapf(
					err, "couldn't check network %s members list for changes", networkID,
				)
			}
			devices = updatedDevices
			if !changed {
				return false, nil
			}
			if !initialized {
				// We just started publishing because a page added a subscription, so there's no need to
				// send the devices list again - that page already has the latest version
				initialized = true
				return false, nil
			}

			// Publish changes
			message, err := replaceDevicesListStream(
				c.Context(), controllerAddress, networkID, auth.Auth{}, h.ztc, h.ztcc, h.dc,
			)
			if err != nil {
				return false, errors.Wrapf(
					err, "couldn't generate turbo streams update for network %s members list", networkID,
				)
			}
			c.Publish(message)
			return false, nil
		})
	}
}

func (h *Handlers) HandleDevicesPost() auth.HTTPHandlerFunc {
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
			return errors.Wrapf(err, "couldn't find controller %s", controllerAddress)
		}
		if err = setMemberAuthorization(
			ctx, *controller, networkID, memberAddress, true, h.ztc,
		); err != nil {
			return errors.Wrapf(
				err, "couldn't authorize network %s member %s", networkID, memberAddress,
			)
		}

		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
			// We send the entire devices list because the content of the devices list partial depends on
			// whether there's at least one device in the network, and this is the simplest solution which
			// handles all edge cases.
			message, err := replaceDevicesListStream(
				c.Request().Context(), controllerAddress, networkID, a, h.ztc, h.ztcc, h.dc,
			)
			if err != nil {
				return errors.Wrapf(
					err, "couldn't generate turbo streams update for network %s members list", networkID,
				)
			}
			return h.r.TurboStream(c.Response(), message)
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf(
			"/networks/%s#/networks/%s/devices/%s", networkID, networkID, memberAddress,
		))
	}
}

// Device

const devicePartial = "networks/device.partial.tmpl"

type DeviceViewData struct {
	Member          Member
	Network         zerotier.ControllerNetwork
	NetworkDNSNamed bool
}

func getDeviceViewData(
	ctx context.Context, controllerAddress, networkID, memberAddress string,
	c *ztc.Client, cc *ztcontrollers.Client, dc *desecc.Client,
) (vd DeviceViewData, err error) {
	controller, err := cc.FindControllerByAddress(ctx, controllerAddress)
	if err != nil {
		return DeviceViewData{}, errors.Wrapf(err, "couldn't find controller %s", controllerAddress)
	}
	if controller == nil {
		return DeviceViewData{}, echo.NewHTTPError(http.StatusNotFound, "controller not found")
	}

	eg, egctx := errgroup.WithContext(ctx)
	var network *zerotier.ControllerNetwork
	var subnameRRsets map[string][]desec.RRset
	eg.Go(func() (err error) {
		network, err = c.GetNetwork(ctx, *controller, networkID)
		return errors.Wrapf(err, "couldn't get network %s", networkID)
	})
	eg.Go(func() (err error) {
		subnameRRsets, err = dc.GetRRsets(ctx)
		return errors.Wrap(err, "couldn't get subname RRsets")
	})
	if err = eg.Wait(); err != nil {
		return DeviceViewData{}, err
	}
	if network == nil {
		return DeviceViewData{}, echo.NewHTTPError(http.StatusNotFound, "zerotier network not found")
	}
	vd.Network = *network

	members, err := getMemberRecords(
		ctx, dc.Config.DomainName, *controller, *network, []string{memberAddress}, subnameRRsets, c,
	)
	if err != nil {
		return DeviceViewData{}, errors.Wrapf(
			err, "couldn't get network %s member %s records", networkID, memberAddress,
		)
	}
	var ok bool
	if vd.Member, ok = members[memberAddress]; !ok {
		return DeviceViewData{}, echo.NewHTTPError(
			http.StatusNotFound, "zerotier network member not found",
		)
	}

	if vd.NetworkDNSNamed, err = checkNamedByDNS(egctx, *network.Name, networkID, dc); err != nil {
		return DeviceViewData{}, errors.Wrapf(
			err, "couldn't check whether network %s has dns-validated name of %s",
			networkID, *network.Name,
		)
	}

	return vd, nil
}

func replaceDeviceStream(
	ctx context.Context, controllerAddress, networkID, memberAddress string, a auth.Auth,
	c *ztc.Client, cc *ztcontrollers.Client, dc *desecc.Client,
) (turbostreams.Message, error) {
	deviceViewData, err := getDeviceViewData(
		ctx, controllerAddress, networkID, memberAddress, c, cc, dc,
	)
	if err != nil {
		return turbostreams.Message{}, errors.Wrapf(
			err, "couldn't get device view data for network %s member %s", networkID, memberAddress,
		)
	}
	return turbostreams.Message{
		Action:   turbostreams.ActionReplace,
		Target:   "/networks/" + networkID + "/devices/" + memberAddress,
		Template: devicePartial,
		Data: map[string]interface{}{
			"Member":          deviceViewData.Member,
			"Network":         deviceViewData.Network,
			"NetworkDNSNamed": deviceViewData.NetworkDNSNamed,
			"Auth":            a,
		},
	}, nil
}

func (h *Handlers) HandleDeviceSub() turbostreams.HandlerFunc {
	return func(c *turbostreams.Context) error {
		// Parse params
		networkID := c.Param("id")
		controllerAddress := ztc.GetControllerAddress(networkID)
		memberAddress := c.Param("address")

		// Run queries
		ctx := c.Context()
		controller, err := h.ztcc.FindControllerByAddress(ctx, controllerAddress)
		if err != nil {
			return errors.Wrapf(err, "couldn't find controller %s", controllerAddress)
		}
		member, err := h.ztc.GetNetworkMember(ctx, *controller, networkID, memberAddress)
		if err != nil {
			return errors.Wrapf(err, "couldn't get network %s member %s", networkID, memberAddress)
		}
		if member == nil {
			return errors.Errorf("network %s member %s not found", networkID, memberAddress)
		}

		// Allow subscription
		return nil
	}
}

func checkNetwork(
	ctx context.Context, controllerAddress, networkID string, prevNetwork zerotier.ControllerNetwork,
	c *ztc.Client, cc *ztcontrollers.Client,
) (changed bool, updatedNetwork zerotier.ControllerNetwork, err error) {
	controller, err := cc.FindControllerByAddress(ctx, controllerAddress)
	if err != nil {
		return false, updatedNetwork, errors.Wrapf(
			err, "couldn't find controller %s", controllerAddress,
		)
	}
	member, err := c.GetNetwork(ctx, *controller, networkID)
	if err != nil {
		return false, updatedNetwork, errors.Wrapf(err, "couldn't get network %s", networkID)
	}
	updatedNetwork = *member
	sixplaneChanged := prevNetwork.V6AssignMode == nil ||
		prevNetwork.V6AssignMode.N6plane == nil ||
		*updatedNetwork.V6AssignMode.N6plane != *prevNetwork.V6AssignMode.N6plane
	rfc4193Changed := prevNetwork.V6AssignMode == nil ||
		prevNetwork.V6AssignMode.Rfc4193 == nil ||
		*updatedNetwork.V6AssignMode.Rfc4193 != *prevNetwork.V6AssignMode.Rfc4193
	if !sixplaneChanged && !rfc4193Changed {
		prevNetwork.V6AssignMode = updatedNetwork.V6AssignMode
		return false, prevNetwork, nil
	}
	return true, updatedNetwork, nil
}

func checkDevice(
	ctx context.Context, controllerAddress, networkID, memberAddress string,
	prevDevice zerotier.ControllerNetworkMember,
	c *ztc.Client, cc *ztcontrollers.Client,
) (changed bool, updatedDevice zerotier.ControllerNetworkMember, err error) {
	controller, err := cc.FindControllerByAddress(ctx, controllerAddress)
	if err != nil {
		return false, updatedDevice, errors.Wrapf(err, "couldn't find controller %s", controllerAddress)
	}
	member, err := c.GetNetworkMember(ctx, *controller, networkID, memberAddress)
	if err != nil {
		return false, updatedDevice, errors.Wrapf(
			err, "couldn't get network %s member %s", networkID, memberAddress,
		)
	}
	updatedDevice = *member
	revisionChanged := prevDevice.Revision == nil ||
		*updatedDevice.Revision != *prevDevice.Revision
	// TODO: do we need to check whether the IP assignments list has changed?
	if !revisionChanged {
		return false, prevDevice, nil
	}
	return true, updatedDevice, nil
}

func (h *Handlers) HandleDevicePub() turbostreams.HandlerFunc {
	t := devicesListPartial
	h.r.MustHave(t)
	return func(c *turbostreams.Context) error {
		// Make change trackers
		initialized := false
		var device zerotier.ControllerNetworkMember
		var network zerotier.ControllerNetwork

		// Parse params
		networkID := c.Param("id")
		controllerAddress := ztc.GetControllerAddress(networkID)
		memberAddress := c.Param("address")

		// Publish periodically
		const pubInterval = 5 * time.Second
		return handling.RepeatImmediate(c.Context(), pubInterval, func() (done bool, err error) {
			// Check for changes
			networkChanged, updatedNetwork, err := checkNetwork(
				c.Context(), controllerAddress, networkID, network, h.ztc, h.ztcc,
			)
			if err != nil {
				return false, errors.Wrapf(
					err, "couldn't check network %s for changes", networkID,
				)
			}
			network = updatedNetwork
			deviceChanged, updatedDevice, err := checkDevice(
				c.Context(), controllerAddress, networkID, memberAddress, device, h.ztc, h.ztcc,
			)
			if err != nil {
				return false, errors.Wrapf(
					err, "couldn't check network %s member %s for changes", networkID, memberAddress,
				)
			}
			device = updatedDevice

			if !deviceChanged && !networkChanged {
				return false, nil
			}
			if !initialized {
				// We just started publishing because a page added a subscription, so there's no need to
				// send the devices list again - that page already has the latest version
				initialized = true
				return false, nil
			}

			// Publish changes
			message, err := replaceDeviceStream(
				c.Context(), controllerAddress, networkID, memberAddress, auth.Auth{}, h.ztc, h.ztcc, h.dc,
			)
			if err != nil {
				return false, errors.Wrapf(
					err, "couldn't generate turbo streams update for network %s member %s",
					networkID, memberAddress,
				)
			}
			c.Publish(message)
			return false, nil
		})
	}
}

// Device Authorization

func setMemberAuthorization(
	ctx context.Context, controller ztcontrollers.Controller, networkID, memberAddress string,
	authorized bool, c *ztc.Client,
) error {
	auth := authorized
	if err := c.UpdateMember(
		ctx, controller, networkID, memberAddress,
		zerotier.SetControllerNetworkMemberJSONRequestBody{Authorized: &auth},
	); err != nil {
		return errors.Wrapf(err, "couldn't update network %s member %s", networkID, memberAddress)
	}
	if authorized {
		// We might've added a new network member, so we should invalidate the cache
		c.Cache.UnsetNetworkMembersByID(networkID)
	}
	return nil
}

func (h *Handlers) HandleDeviceAuthorizationPost() auth.HTTPHandlerFunc {
	t := devicesListPartial
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		networkID := c.Param("id")
		controllerAddress := ztc.GetControllerAddress(networkID)
		memberAddress := c.Param("address")
		authorization := strings.ToLower(c.FormValue("authorization")) == checkboxTrueValue

		// Run queries
		ctx := c.Request().Context()
		controller, err := h.ztcc.FindControllerByAddress(ctx, controllerAddress)
		if err != nil {
			return errors.Wrapf(err, "couldn't find controller %s", controllerAddress)
		}
		if err = setMemberAuthorization(
			ctx, *controller, networkID, memberAddress, authorization, h.ztc,
		); err != nil {
			return errors.Wrapf(
				err, "couldn't update authorization on network %s for member %s", networkID, memberAddress,
			)
		}

		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
			// We send the entire devices list because we already have to look up roughly the same
			// amount of data to give the device partial, and it's probably not worth the additional code
			// complexity to try to only look up the data for this device in order to send a smaller
			// HTTP response payload.
			message, err := replaceDevicesListStream(
				c.Request().Context(), controllerAddress, networkID, a, h.ztc, h.ztcc, h.dc,
			)
			if err != nil {
				return errors.Wrapf(
					err, "couldn't generate turbo streams update for network %s members list", networkID,
				)
			}
			return h.r.TurboStream(c.Response(), message)
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf(
			"/networks/%s#/networks/%s/devices/%s", networkID, networkID, memberAddress,
		))
	}
}

// Device Naming

func confirmMemberNameManageable(
	ctx context.Context, network zerotier.ControllerNetwork, memberName string, dc *desecc.Client,
) (memberSubname string, err error) {
	networkName := *network.Name
	named, err := checkNamedByDNS(ctx, networkName, *network.Id, dc)
	if err != nil {
		return "", errors.Wrapf(
			err, "couldn't check whether network %s has dns-validated name of %s",
			*network.Id, *network.Name,
		)
	}
	if !named {
		return "", errors.Errorf("network does not have a valid domain name for naming members")
	}

	// TODO: check whether the member name was already allocated!

	fqdn := fmt.Sprintf("%s.d.%s", memberName, networkName)
	return strings.TrimSuffix(fqdn, fmt.Sprintf(".%s", dc.Config.DomainName)), nil
}

func setMemberName(
	ctx context.Context, controller ztcontrollers.Controller, networkID string,
	memberAddress, memberName string, c *ztc.Client, dc *desecc.Client,
) error {
	network, err := c.GetNetwork(ctx, controller, networkID)
	if err != nil {
		return errors.Wrapf(err, "couldn't get network %s", networkID)
	}
	memberSubname, err := confirmMemberNameManageable(ctx, *network, memberName, dc)
	if err != nil {
		return errors.Wrapf(err, "network %s can't manage member name %s", networkID, memberName)
	}
	member, err := c.GetNetworkMember(ctx, controller, networkID, memberAddress)
	if err != nil {
		return errors.Wrapf(err, "couldn't get network %s member %s", networkID, memberAddress)
	}

	ipAddresses, _, err := calculateIPAddresses(*network.Id, *network.V6AssignMode, *member)
	if err != nil {
		return errors.Wrapf(
			err, "couldn't calculate IP addresses for network %s member %s", networkID, memberAddress,
		)
	}
	ipv4Addresses, ipv6Addresses, err := splitIPAddresses(ipAddresses)
	if err != nil {
		return errors.Wrapf(
			err, "found invalid IP address for network %s member %s", networkID, memberAddress,
		)
	}

	// TODO: use bulk Update
	ttl := int(c.Config.DNS.DeviceTTL)
	rrsets := []desec.RRset{
		{
			Subname: memberSubname,
			Type:    "AAAA",
			Ttl:     &ttl,
			Records: ipv6Addresses,
		},
		{
			Subname: memberSubname,
			Type:    "A",
			Ttl:     &ttl,
			Records: ipv4Addresses,
		},
	}
	if _, err := dc.UpsertRRsets(ctx, rrsets...); err != nil {
		return errors.Wrapf(
			err, "couldn't upsert AAAA and/or A records of %s for network %s member %s",
			memberSubname, networkID, memberAddress,
		)
	}
	return nil
}

func unsetMemberName(
	ctx context.Context, controller ztcontrollers.Controller, networkID string,
	memberName string, c *ztc.Client, dc *desecc.Client,
) error {
	network, err := c.GetNetwork(ctx, controller, networkID)
	if err != nil {
		return errors.Wrapf(err, "couldn't get network %s", networkID)
	}
	memberSubname, err := confirmMemberNameManageable(ctx, *network, memberName, dc)
	if err != nil {
		return errors.Wrapf(err, "network %s can't manage member name %s", networkID, memberName)
	}

	// TODO: use bulk deletion
	deletionKeys := []desecc.RRsetKey{
		{
			Subname: memberSubname,
			Type:    "AAAA",
		},
		{
			Subname: memberSubname,
			Type:    "A",
		},
	}
	if err := dc.DeleteRRsets(ctx, deletionKeys...); err != nil {
		return errors.Wrapf(
			err, "couldn't delete AAAA and A records of %s in network %s member",
			memberSubname, networkID,
		)
	}
	return nil
}

func (h *Handlers) HandleDeviceNamePost() auth.HTTPHandlerFunc {
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
			return errors.Wrapf(err, "couldn't find controller %s", controllerAddress)
		}

		switch setName {
		default:
			if err = setMemberName(
				ctx, *controller, networkID, memberAddress, setName, h.ztc, h.dc,
			); err != nil {
				return errors.Wrapf(
					err, "couldn't set name of network %s member %s to %s", networkID, memberAddress, setName,
				)
			}
		case "":
			nameToUnset := c.FormValue("unset-name")
			if err = unsetMemberName(
				ctx, *controller, networkID, nameToUnset, h.ztc, h.dc,
			); err != nil {
				return errors.Wrapf(
					err, "couldn't unset name %s of network %s member %s", setName, networkID, memberAddress,
				)
			}
		}

		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
			// We send the entire devices list because we already have to look up roughly the same
			// amount of data to give the device partial, and it's probably not worth the additional code
			// complexity to try to only look up the data for this device in order to send a smaller
			// HTTP response payload.
			message, err := replaceDeviceStream(
				c.Request().Context(), controllerAddress, networkID, memberAddress, a, h.ztc, h.ztcc, h.dc,
			)
			if err != nil {
				return errors.Wrapf(
					err, "couldn't generate turbo streams update for network %s member %s",
					networkID, memberAddress,
				)
			}
			return h.r.TurboStream(c.Response(), message)
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf(
			"/networks/%s#/networks/%s/device/%s", networkID, networkID, memberAddress,
		))
	}
}

// Device IP Addresses

func setDeviceIPAddresses(
	ctx context.Context, controller ztcontrollers.Controller, networkID, memberAddress string,
	ipAddresses []string, c *ztc.Client,
) error {
	addresses := make([]string, len(ipAddresses))
	for i := range ipAddresses {
		addresses[i] = strings.TrimSpace(ipAddresses[i])
	}
	if err := c.UpdateMember(
		ctx, controller, networkID, memberAddress,
		zerotier.SetControllerNetworkMemberJSONRequestBody{IpAssignments: &addresses},
	); err != nil {
		return errors.Wrapf(err, "couldn't update network %s member %s", networkID, memberAddress)
	}
	return nil
}

func (h *Handlers) HandleDeviceIPPost() auth.HTTPHandlerFunc {
	t := devicePartial
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		networkID := c.Param("id")
		controllerAddress := ztc.GetControllerAddress(networkID)
		memberAddress := c.Param("address")
		formParams, err := c.FormParams()
		if err != nil {
			return errors.Wrap(err, "couldn't parse form params")
		}

		// Run queries
		ctx := c.Request().Context()
		controller, err := h.ztcc.FindControllerByAddress(ctx, controllerAddress)
		if err != nil {
			return errors.Wrapf(err, "couldn't find controller %s", controllerAddress)
		}
		ipAddresses := formParams["existing-addresses"]
		if newAddress := c.FormValue("new-address"); len(newAddress) > 0 {
			ipAddresses = append(ipAddresses, newAddress)
		}
		if err = setDeviceIPAddresses(
			ctx, *controller, networkID, memberAddress, ipAddresses, h.ztc,
		); err != nil {
			return errors.Wrapf(
				err, "couldn't set ip addresses of network %s member %s", networkID, memberAddress,
			)
		}

		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
			// TODO: also broadcast this message over Turbo Streams, and have web browsers subscribe to it
			message, err := replaceDeviceStream(
				ctx, controllerAddress, networkID, memberAddress, a, h.ztc, h.ztcc, h.dc,
			)
			if err != nil {
				return errors.Wrapf(
					err, "couldn't generate turbo streams update for network %s member %s",
					networkID, memberAddress,
				)
			}
			return h.r.TurboStream(c.Response(), message)
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf(
			"/networks/%s#/networks/%s/devices/%s/ip", networkID, networkID, memberAddress,
		))
	}
}

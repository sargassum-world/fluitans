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

// Device Membership Comparison

type StringSet map[string]bool

func NewStringSet(strings []string) StringSet {
	set := make(map[string]bool)
	for _, s := range strings {
		set[s] = true
	}
	return set
}

func (ss StringSet) Contains(set StringSet) bool {
	if ss == nil || set == nil {
		return false
	}
	if len(set) > len(ss) {
		return false
	}

	for s := range set {
		if _, ok := ss[s]; !ok {
			return false
		}
	}
	return true
}

func (ss StringSet) Equals(set StringSet) bool {
	if ss == nil || set == nil {
		return false
	}
	if len(set) != len(ss) {
		return false
	}

	// This might not be the most efficient algorithm, but it's fine for now
	return ss.Contains(set) && set.Contains(ss)
}

// Devices

const devicesListPartial = "networks/devices-list.partial.tmpl"

func replaceDevicesListStream(
	ctx context.Context, controllerAddress, networkID string, a auth.Auth,
	c *ztc.Client, cc *ztcontrollers.Client, dc *desecc.Client,
) (turbostreams.Message, error) {
	networkViewData, err := getNetworkViewData(ctx, controllerAddress, networkID, c, cc, dc)
	if err != nil {
		return turbostreams.Message{}, err
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
			return err
		}
		network, err := h.ztc.GetNetwork(ctx, *controller, networkID)
		if err != nil {
			return err
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
		return false, prevDevices, err
	}
	addresses, err := c.GetNetworkMemberAddresses(ctx, *controller, networkID)
	if err != nil {
		return false, prevDevices, err
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
				return false, err
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
				return false, err
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
			return err
		}
		if err = setMemberAuthorization(
			ctx, *controller, networkID, memberAddress, true, h.ztc,
		); err != nil {
			return err
		}

		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
			// We send the entire devices list because the content of the devices list partial depends on
			// whether there's at least one device in the network, and this is the simplest solution which
			// handles all edge cases.
			replaceStream, err := replaceDevicesListStream(
				c.Request().Context(), controllerAddress, networkID, a, h.ztc, h.ztcc, h.dc,
			)
			if err != nil {
				return err
			}
			return h.r.TurboStream(c.Response(), replaceStream)
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
	DomainNames     []string
	Network         zerotier.ControllerNetwork
	NetworkDNSNamed bool
}

func getDeviceViewData(
	ctx context.Context, address, networkID, memberAddress string,
	c *ztc.Client, cc *ztcontrollers.Client, dc *desecc.Client,
) (vd DeviceViewData, err error) {
	controller, err := cc.FindControllerByAddress(ctx, address)
	if err != nil {
		return DeviceViewData{}, err
	}
	if controller == nil {
		return DeviceViewData{}, echo.NewHTTPError(http.StatusNotFound, "controller not found")
	}

	eg, egctx := errgroup.WithContext(ctx)
	var network *zerotier.ControllerNetwork
	var subnameRRsets map[string][]desec.RRset
	eg.Go(func() (err error) {
		network, err = c.GetNetwork(ctx, *controller, networkID)
		return err
	})
	eg.Go(func() (err error) {
		subnameRRsets, err = dc.GetRRsets(ctx)
		return err
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
		return DeviceViewData{}, err
	}
	var ok bool
	if vd.Member, ok = members[memberAddress]; !ok {
		return DeviceViewData{}, echo.NewHTTPError(
			http.StatusNotFound, "zerotier network member not found",
		)
	}

	if vd.NetworkDNSNamed, err = checkNamedByDNS(egctx, *network.Name, *network.Id, dc); err != nil {
		return DeviceViewData{}, err
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
		return turbostreams.Message{}, err
	}
	return turbostreams.Message{
		Action:   turbostreams.ActionReplace,
		Target:   "/networks/" + networkID + "/devices/" + memberAddress,
		Template: devicePartial,
		Data: map[string]interface{}{
			"Member":          deviceViewData.Member,
			"DomainNames":     deviceViewData.DomainNames,
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
			return err
		}
		member, err := h.ztc.GetNetworkMember(ctx, *controller, networkID, memberAddress)
		if err != nil {
			return err
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
		return false, updatedNetwork, err
	}
	member, err := c.GetNetwork(ctx, *controller, networkID)
	if err != nil {
		return false, updatedNetwork, err
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
		return false, updatedDevice, err
	}
	member, err := c.GetNetworkMember(ctx, *controller, networkID, memberAddress)
	if err != nil {
		return false, updatedDevice, err
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
				return false, err
			}
			network = updatedNetwork
			deviceChanged, updatedDevice, err := checkDevice(
				c.Context(), controllerAddress, networkID, memberAddress, device, h.ztc, h.ztcc,
			)
			if err != nil {
				return false, err
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
				return false, err
			}
			c.Publish(message)
			return false, nil
		})
	}
}

// Authorization

func setMemberAuthorization(
	ctx context.Context, controller ztcontrollers.Controller, networkID, memberAddress string,
	authorized bool, c *ztc.Client,
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
			return err
		}
		if err = setMemberAuthorization(
			ctx, *controller, networkID, memberAddress, authorization, h.ztc,
		); err != nil {
			return err
		}

		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
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
			return h.r.TurboStream(c.Response(), replaceStream)
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf(
			"/networks/%s#/networks/%s/devices/%s", networkID, networkID, memberAddress,
		))
	}
}

// Naming

func confirmMemberNameManageable(
	ctx context.Context, controller ztcontrollers.Controller, networkID string,
	memberAddress, memberName string, c *ztc.Client, dc *desecc.Client,
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
	memberAddress, memberName string, c *ztc.Client, dc *desecc.Client,
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
	memberAddress, memberName string, c *ztc.Client, dc *desecc.Client,
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
		if turbostreams.Accepted(c.Request().Header) {
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
			return h.r.TurboStream(c.Response(), replaceStream)
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf(
			"/networks/%s#/networks/%s/device/%s", networkID, networkID, memberAddress,
		))
	}
}

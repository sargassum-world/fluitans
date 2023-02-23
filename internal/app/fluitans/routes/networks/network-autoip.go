package networks

import (
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/turbostreams"
	"go4.org/netipx"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	ztc "github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/zerotier"
)

// Network IP Assignment Pools

type AssignmentPool struct {
	Range       netipx.IPRange
	Enabled     bool
	ExactPrefix bool
	Prefix      netip.Prefix
}

func sanitizeAssignmentPoolRange(poolRange netipx.IPRange) netipx.IPRange {
	poolRangeFrom := poolRange.From()
	poolRangeTo := poolRange.To()
	if !poolRangeFrom.Is4() || !poolRangeTo.Is4() {
		return poolRange
	}

	from := poolRangeFrom.As4()
	const lastOctetIndex = len(from) - 1
	if from[lastOctetIndex] == 0 {
		from[lastOctetIndex] += 1
	}
	poolRangeFrom = netip.AddrFrom4(from)

	to := poolRangeTo.As4()
	const broadcastSuffix = 255
	if to[lastOctetIndex] == broadcastSuffix {
		to[lastOctetIndex] -= 1
	}
	poolRangeTo = netip.AddrFrom4(to)

	if poolRangeFrom.Compare(poolRangeTo) <= 0 {
		poolRange = netipx.IPRangeFrom(poolRangeFrom, poolRangeTo)
	}
	return poolRange
}

func parseAssignmentPools(
	rawRoutes []zerotier.Route, rawPools []zerotier.IpAssignmentPool,
) (pools []AssignmentPool, err error) {
	keyedPools := make(map[netipx.IPRange]AssignmentPool)
	poolRanges := make([]netipx.IPRange, 0, len(rawRoutes)+len(rawPools))
	for _, rawRoute := range rawRoutes {
		if rawRoute.Target == nil {
			return nil, errors.New("couldn't find target in zerotier managed route")
		}
		prefix, err := netip.ParsePrefix(*rawRoute.Target)
		if err != nil {
			return nil, errors.New("couldn't parse zerotier managed route target")
		}
		if rawRoute.Via != nil && len(*rawRoute.Via) > 0 {
			// This route is routed via something else, so we shouldn't assign IP addresses to it
			continue
		}

		poolRange := sanitizeAssignmentPoolRange(netipx.RangeOfPrefix(prefix))
		if _, ok := keyedPools[poolRange]; ok {
			// We already have the assignment pool
			continue
		}
		keyedPools[poolRange] = AssignmentPool{
			Range:       poolRange,
			ExactPrefix: true,
			Prefix:      prefix,
		}
		poolRanges = append(poolRanges, poolRange)
	}
	for _, rawRange := range rawPools {
		rangeStart, err := netip.ParseAddr(*rawRange.IpRangeStart)
		if err != nil {
			return nil, errors.New("couldn't parse zerotier ip assignment pool starting address")
		}
		rangeEnd, err := netip.ParseAddr(*rawRange.IpRangeEnd)
		if err != nil {
			return nil, errors.New("couldn't parse zerotier ip assignment pool ending address")
		}
		poolRange := netipx.IPRangeFrom(rangeStart, rangeEnd)
		if _, ok := keyedPools[poolRange]; ok {
			// We already have the assignment pool
			pool := keyedPools[poolRange]
			pool.Enabled = true
			keyedPools[poolRange] = pool
			continue
		}
		pool := AssignmentPool{
			Range:   poolRange,
			Enabled: true,
		}
		pool.Prefix, pool.ExactPrefix = poolRange.Prefix()
		keyedPools[poolRange] = pool
		poolRanges = append(poolRanges, poolRange)
	}

	pools = make([]AssignmentPool, len(poolRanges))
	for i, poolRange := range poolRanges {
		pools[i] = keyedPools[poolRange]
	}
	return pools, nil
}

// Network IP Auto-Assignment Pools

const autoIPPoolsPartial = "networks/network-autoip-pools.partial.tmpl"

func replaceAutoIPPoolsStream(
	id string, network *zerotier.ControllerNetwork, assignmentPools []AssignmentPool, a auth.Auth,
) turbostreams.Message {
	return turbostreams.Message{
		Action:   turbostreams.ActionReplace,
		Target:   "/networks/" + id + "/autoip/pools",
		Template: autoIPPoolsPartial,
		Data: map[string]interface{}{
			"Network":         network,
			"AssignmentPools": assignmentPools,
			"Auth":            a,
		},
	}
}

func setNetworkAutoIPPools(
	ctx context.Context, controller ztcontrollers.Controller,
	id string, rawRanges []string, c *ztc.Client,
) (*zerotier.ControllerNetwork, error) {
	pools := make([]zerotier.IpAssignmentPool, len(rawRanges))
	for i := range rawRanges {
		sanitizedRange := strings.ReplaceAll(rawRanges[i], " ", "")
		ipRange, err := netipx.ParseIPRange(sanitizedRange)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't parse ip range %s", sanitizedRange)
		}
		rangeStart := ipRange.From().String()
		rangeEnd := ipRange.To().String()
		pools[i].IpRangeStart = &rangeStart
		pools[i].IpRangeEnd = &rangeEnd
	}
	network, err := c.UpdateNetwork(
		ctx, controller, id, zerotier.SetControllerNetworkJSONRequestBody{IpAssignmentPools: &pools},
	)
	if err != nil {
		return nil, err
	}
	return network, nil
}

func (h *Handlers) HandleNetworkAutoIPPoolsPost() auth.HTTPHandlerFunc {
	t := autoIPPoolsPartial
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id := c.Param("id")
		address := ztc.GetControllerAddress(id)
		formParams, err := c.FormParams()
		if err != nil {
			return errors.Wrap(err, "couldn't parse form params")
		}

		// Run queries
		ctx := c.Request().Context()
		controller, err := h.ztcc.FindControllerByAddress(ctx, address)
		if err != nil {
			return err
		}
		ranges := formParams["existing-pools"]
		if newPool := c.FormValue("new-pool"); len(newPool) > 0 {
			ranges = append(ranges, newPool)
		}
		network, err := setNetworkAutoIPPools(ctx, *controller, id, ranges, h.ztc)
		if err != nil {
			return err
		}

		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
			// TODO: also broadcast this message over Turbo Streams, and have web browsers subscribe to it
			assignmentPools, err := parseAssignmentPools(*network.Routes, *network.IpAssignmentPools)
			if err != nil {
				return err
			}
			return h.r.TurboStream(
				c.Response(),
				replaceAutoIPPoolsStream(id, network, assignmentPools, a),
			)
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf(
			"/networks/%s#/networks/%s/autoip/pools", id, id,
		))
	}
}

// Network IP Auto-Assignment Modes

func setNetworkAutoIPv6Modes(
	ctx context.Context, controller ztcontrollers.Controller,
	id string, modes zerotier.V6AssignMode, c *ztc.Client,
) (*zerotier.ControllerNetwork, error) {
	network, err := c.UpdateNetwork(
		ctx, controller, id, zerotier.SetControllerNetworkJSONRequestBody{V6AssignMode: &modes},
	)
	if err != nil {
		return nil, err
	}
	return network, nil
}

const checkboxTrueValue = "true"

func (h *Handlers) HandleNetworkAutoIPv6ModesPost() auth.HTTPHandlerFunc {
	t := "networks/network-autoip-v6modes.partial.tmpl"
	h.r.MustHave(t)
	h.r.MustHave(autoIPPoolsPartial)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id := c.Param("id")
		address := ztc.GetControllerAddress(id)
		sixplaneEnabled := strings.ToLower(c.FormValue("sixplane")) == checkboxTrueValue
		rfc4193Enabled := strings.ToLower(c.FormValue("rfc4193")) == checkboxTrueValue
		zerotierEnabled := strings.ToLower(c.FormValue("zerotier")) == checkboxTrueValue

		// Run queries
		ctx := c.Request().Context()
		controller, err := h.ztcc.FindControllerByAddress(ctx, address)
		if err != nil {
			return err
		}
		network, err := setNetworkAutoIPv6Modes(
			ctx, *controller, id, zerotier.V6AssignMode{
				N6plane: &sixplaneEnabled,
				Rfc4193: &rfc4193Enabled,
				Zt:      &zerotierEnabled,
			}, h.ztc,
		)
		if err != nil {
			return err
		}

		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
			// TODO: also broadcast this message over Turbo Streams, and have web browsers subscribe to it
			assignmentPools, err := parseAssignmentPools(*network.Routes, *network.IpAssignmentPools)
			if err != nil {
				return err
			}
			return h.r.TurboStream(
				c.Response(),
				turbostreams.Message{
					Action:   turbostreams.ActionReplace,
					Target:   "/networks/" + id + "/autoip/v6-modes",
					Template: t,
					Data: map[string]interface{}{
						"Network": network,
						"Auth":    a,
					},
				},
				replaceAutoIPPoolsStream(id, network, assignmentPools, a),
			)
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf(
			"/networks/%s#/networks/%s/autoip/v6-modes", id, id,
		))
	}
}

func setNetworkAutoIPv4Modes(
	ctx context.Context, controller ztcontrollers.Controller,
	id string, modes zerotier.V4AssignMode, c *ztc.Client,
) (*zerotier.ControllerNetwork, error) {
	network, err := c.UpdateNetwork(
		ctx, controller, id, zerotier.SetControllerNetworkJSONRequestBody{V4AssignMode: &modes},
	)
	if err != nil {
		return nil, err
	}
	return network, nil
}

func (h *Handlers) HandleNetworkAutoIPv4ModesPost() auth.HTTPHandlerFunc {
	t := "networks/network-autoip-v4modes.partial.tmpl"
	h.r.MustHave(t)
	h.r.MustHave(autoIPPoolsPartial)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id := c.Param("id")
		address := ztc.GetControllerAddress(id)
		zerotierEnabled := strings.ToLower(c.FormValue("zerotier")) == "true"

		// Run queries
		ctx := c.Request().Context()
		controller, err := h.ztcc.FindControllerByAddress(ctx, address)
		if err != nil {
			return err
		}
		network, err := setNetworkAutoIPv4Modes(
			ctx, *controller, id, zerotier.V4AssignMode{
				Zt: &zerotierEnabled,
			}, h.ztc,
		)
		if err != nil {
			return err
		}

		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
			// TODO: also broadcast this message over Turbo Streams, and have web browsers subscribe to it
			assignmentPools, err := parseAssignmentPools(*network.Routes, *network.IpAssignmentPools)
			if err != nil {
				return err
			}
			return h.r.TurboStream(
				c.Response(),
				turbostreams.Message{
					Action:   turbostreams.ActionReplace,
					Target:   "/networks/" + id + "/autoip/v4-modes",
					Template: t,
					Data: map[string]interface{}{
						"Network": network,
						"Auth":    a,
					},
				},
				replaceAutoIPPoolsStream(id, network, assignmentPools, a),
			)
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf(
			"/networks/%s#/networks/%s/autoip/v4-modes", id, id,
		))
	}
}

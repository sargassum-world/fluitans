package networks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/turbostreams"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/client"
	desecc "github.com/sargassum-world/fluitans/internal/clients/desec"
	ztc "github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/desec"
	"github.com/sargassum-world/fluitans/pkg/zerotier"
)

// Network DNS

func identifyNetworkAliases(
	networkID, confirmedSubname string, txtRecords map[string][]string,
) []string {
	subnames := make([]string, 0, len(txtRecords))
	for subname := range txtRecords {
		subnames = append(subnames, subname)
	}
	sort.Strings(subnames)

	aliases := make([]string, 0, len(txtRecords))
	for _, subname := range subnames {
		if subname == confirmedSubname {
			continue
		}

		if id, has := client.GetNetworkID(txtRecords[subname]); has && id == networkID {
			aliases = append(aliases, subname)
		}
	}
	return aliases
}

type NetworkDNS struct {
	Named            bool
	Aliases          []string
	DeviceSubdomains map[string]client.Subdomain
	OtherSubdomains  []client.Subdomain
}

func getNetworkDNSRecords(
	ctx context.Context, networkID, networkName string, subnameRRsets map[string][]desec.RRset,
	c *ztc.Client, cc *ztcontrollers.Client, dc *desecc.Client,
) (networkDNS NetworkDNS, err error) {
	if !client.NetworkNamedByDNS(networkID, networkName, dc.Config.DomainName, subnameRRsets) {
		return NetworkDNS{}, nil
	}
	networkDNS.Named = true

	txtRecords, err := client.GetRecordsOfType(subnameRRsets, "TXT")
	if err != nil {
		return NetworkDNS{}, err
	}
	confirmedSubname := strings.TrimSuffix(networkName, "."+dc.Config.DomainName)
	networkDNS.Aliases = identifyNetworkAliases(networkID, confirmedSubname, txtRecords)
	aliases := make(map[string]bool, len(networkDNS.Aliases))
	for _, alias := range networkDNS.Aliases {
		aliases[alias] = true
	}

	subdomains, err := client.GetSubdomains(ctx, subnameRRsets, dc, c, cc)
	if err != nil {
		return NetworkDNS{}, err
	}
	networkSubname := strings.TrimSuffix(networkName, "."+dc.Config.DomainName)
	networkDNS.DeviceSubdomains = make(map[string]client.Subdomain)
	for _, subdomain := range subdomains {
		if subdomain.Subname != networkSubname && !strings.HasSuffix(
			subdomain.Subname, "."+networkSubname,
		) && !aliases[subdomain.Subname] {
			// Subdomain is unrelated to this network
			continue
		}

		if strings.HasSuffix(subdomain.Subname, ".d."+networkSubname) {
			// Subdomain is for a device
			networkDNS.DeviceSubdomains[subdomain.Subname] = subdomain
			continue
		}

		networkDNS.OtherSubdomains = append(networkDNS.OtherSubdomains, subdomain)
	}
	return networkDNS, nil
}

type NetworkViewData struct {
	Controller       ztcontrollers.Controller
	Network          zerotier.ControllerNetwork
	Members          []client.Member
	AssignmentPools  []AssignmentPool
	JSONPrintedRules string
	DomainName       string
	NetworkDNS       NetworkDNS
}

func printJSONRules(rawRules []map[string]interface{}) (string, error) {
	rules, err := json.MarshalIndent(rawRules, "", "  ")
	if err != nil {
		return "", err
	}
	return string(rules), err
}

func getNetworkViewData(
	ctx context.Context, address, id string,
	c *ztc.Client, cc *ztcontrollers.Client, dc *desecc.Client,
) (vd NetworkViewData, err error) {
	controller, err := cc.FindControllerByAddress(ctx, address)
	if err != nil {
		return NetworkViewData{}, err
	}
	if controller == nil {
		return NetworkViewData{}, echo.NewHTTPError(http.StatusNotFound, "controller not found")
	}
	vd.Controller = *controller

	eg, egctx := errgroup.WithContext(ctx)
	var network *zerotier.ControllerNetwork
	var memberAddresses []string
	var subnameRRsets map[string][]desec.RRset
	eg.Go(func() (err error) {
		network, memberAddresses, err = c.GetNetworkInfo(egctx, *controller, id)
		return err
	})
	eg.Go(func() (err error) {
		subnameRRsets, err = dc.GetRRsets(egctx)
		return err
	})
	if err = eg.Wait(); err != nil {
		return NetworkViewData{}, err
	}
	if network == nil {
		return NetworkViewData{}, echo.NewHTTPError(http.StatusNotFound, "zerotier network not found")
	}
	vd.Network = *network
	if vd.AssignmentPools, err = parseAssignmentPools(
		*network.Routes, *network.IpAssignmentPools,
	); err != nil {
		return NetworkViewData{}, err
	}
	if vd.JSONPrintedRules, err = printJSONRules(*network.Rules); err != nil {
		return NetworkViewData{}, err
	}

	eg, egctx = errgroup.WithContext(ctx)
	eg.Go(func() (err error) {
		members, err := client.GetMemberRecords(
			egctx, dc.Config.DomainName, *controller, *network, memberAddresses, subnameRRsets, c,
		)
		_, vd.Members = client.SortNetworkMembers(members)
		return err
	})
	eg.Go(func() (err error) {
		vd.NetworkDNS, err = getNetworkDNSRecords(
			egctx, *network.Id, *network.Name, subnameRRsets, c, cc, dc,
		)
		return err
	})
	if err := eg.Wait(); err != nil {
		return NetworkViewData{}, err
	}

	vd.DomainName = dc.Config.DomainName
	return vd, nil
}

// Network

func (h *Handlers) HandleNetworkGet() auth.HTTPHandlerFunc {
	t := "networks/network.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id := c.Param("id")
		address := ztc.GetControllerAddress(id)

		// Run queries
		networkViewData, err := getNetworkViewData(
			c.Request().Context(), address, id, h.ztc, h.ztcc, h.dc,
		)
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, networkViewData, a)
	}
}

func (h *Handlers) HandleNetworkPost() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse params
		id := c.Param("id")
		address := ztc.GetControllerAddress(id)
		state := c.FormValue("state")

		// Run queries
		ctx := c.Request().Context()
		switch state {
		default:
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"invalid network state %s", state,
			))
		case "deleted":
			controller, err := h.ztcc.FindControllerByAddress(ctx, address)
			if err != nil {
				return err
			}
			if err = h.ztc.DeleteNetwork(ctx, *controller, id, h.ztcc); err != nil {
				// TODO: add a tombstone to the TXT RRset?
				return err
			}
			h.tsh.Cancel("/networks/" + id + "/devices")

			// Redirect user
			return c.Redirect(http.StatusSeeOther, "/networks")
		}
	}
}

// Network Name

func nameNetwork(
	ctx context.Context, controller ztcontrollers.Controller, id string, name string,
	c *ztc.Client, dc *desecc.Client,
) (*zerotier.ControllerNetwork, error) {
	if len(name) == 0 {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "cannot remove name from network")
	}

	// Check to see if the network was already named by DNS
	fqdn := name + "." + dc.Config.DomainName
	txtRRset, err := dc.GetRRset(ctx, name, "TXT")
	if err != nil {
		return nil, errors.Wrapf(
			err, "couldn't check cache for DNS TXT RRset at %s for network %s", fqdn, id,
		)
	}
	if txtRRset != nil {
		if _, hasID := client.GetNetworkID(txtRRset.Records); hasID {
			return nil, echo.NewHTTPError(
				http.StatusBadRequest, "name is already used by another network",
			)
		}
	}

	ttl := c.Config.DNS.NetworkTTL
	if _, err := dc.CreateRRset(
		ctx, name, "TXT", ttl, []string{client.MakeNetworkIDRecord(id)},
	); err != nil {
		// TODO: if a TXT RRset already exists but doesn't have the ID, just append a
		// zerotier-net-id=... record (but we should have a global lock on a get-and-patch to avoid
		// data races)
		// TODO: if the returned error code was an HTTP error, preserve the status code
		return nil, errors.Wrapf(
			err, "couldn't create a DNS TXT RRset at %s for network %s", fqdn, id,
		)
	}
	return c.UpdateNetwork(
		ctx, controller, id, zerotier.SetControllerNetworkJSONRequestBody{Name: &fqdn},
	)
}

func (h *Handlers) HandleNetworkNamePost() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse params
		id := c.Param("id")
		address := ztc.GetControllerAddress(id)

		// Run queries
		ctx := c.Request().Context()
		controller, err := h.ztcc.FindControllerByAddress(ctx, address)
		if err != nil {
			return err
		}
		if _, err = nameNetwork(ctx, *controller, id, c.FormValue("name"), h.ztc, h.dc); err != nil {
			return err
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s", id))
	}
}

// Network-Managed Routes

func setNetworkRoutes(
	ctx context.Context, controller ztcontrollers.Controller,
	id string, targets []string, c *ztc.Client,
) (*zerotier.ControllerNetwork, error) {
	routes := make([]zerotier.Route, len(targets))
	for i := range targets {
		sanitized := strings.TrimSpace(targets[i])
		routes[i].Target = &sanitized
	}
	network, err := c.UpdateNetwork(
		ctx, controller, id, zerotier.SetControllerNetworkJSONRequestBody{Routes: &routes},
	)
	if err != nil {
		return nil, err
	}
	return network, nil
}

func (h *Handlers) HandleNetworkRoutesPost() auth.HTTPHandlerFunc {
	t := "networks/network-routes.partial.tmpl"
	h.r.MustHave(t)
	tPools := "networks/network-autoip-pools.partial.tmpl"
	h.r.MustHave(tPools)
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
		targets := formParams["existing-targets"]
		if newTarget := c.FormValue("new-target"); len(newTarget) > 0 {
			targets = append(targets, newTarget)
		}
		network, err := setNetworkRoutes(ctx, *controller, id, targets, h.ztc)
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
					Target:   "/networks/" + id + "/routes",
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
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s#/networks/%s/routes", id, id))
	}
}

// Network Rules

func setNetworkRules(
	ctx context.Context, controller ztcontrollers.Controller,
	id string, jsonRules string, c *ztc.Client,
) (*zerotier.ControllerNetwork, error) {
	rules := make([]map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(jsonRules), &rules); err != nil {
		return nil, err
	}
	network, err := c.UpdateNetwork(
		ctx, controller, id, zerotier.SetControllerNetworkJSONRequestBody{Rules: &rules},
	)
	if err != nil {
		return nil, err
	}
	return network, nil
}

func (h *Handlers) HandleNetworkRulesPost() auth.HTTPHandlerFunc {
	t := "networks/network-rules.partial.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id := c.Param("id")
		address := ztc.GetControllerAddress(id)

		// Run queries
		ctx := c.Request().Context()
		controller, err := h.ztcc.FindControllerByAddress(ctx, address)
		if err != nil {
			return err
		}
		network, err := setNetworkRules(
			ctx, *controller, id, c.FormValue("rules"), h.ztc,
		)
		if err != nil {
			return err
		}

		// Render Turbo Stream if accepted
		if turbostreams.Accepted(c.Request().Header) {
			rules, err := printJSONRules(*network.Rules)
			if err != nil {
				return err
			}
			// TODO: also broadcast this message over Turbo Streams, and have web browsers subscribe to it
			return h.r.TurboStream(c.Response(), turbostreams.Message{
				Action:   turbostreams.ActionReplace,
				Target:   "/networks/" + id + "/rules",
				Template: t,
				Data: map[string]interface{}{
					"Network":          network,
					"Auth":             a,
					"JSONPrintedRules": rules,
				},
			})
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s#/networks/%s/rules", id, id))
	}
}

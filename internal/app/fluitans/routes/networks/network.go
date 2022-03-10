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
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/client"
	desecc "github.com/sargassum-world/fluitans/internal/clients/desec"
	ztc "github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/desec"
	"github.com/sargassum-world/fluitans/pkg/zerotier"
)

// Network Members & Member DNS

func getRecordsOfType(
	subnameRRsets map[string][]desec.RRset, rrsetType string,
) (map[string][]string, error) {
	records := make(map[string][]string)
	// Look up potential domain names of network members
	for subname, rrsets := range subnameRRsets {
		filtered := desecc.FilterAndSortRRsets(rrsets, []string{rrsetType})
		if len(filtered) > 1 {
			return nil, errors.Errorf("unexpected number of RRsets for record")
		}
		if len(filtered) == 1 {
			records[subname] = filtered[0].Records
		}
	}
	return records, nil
}

func identifyMemberDomainNames(
	zoneDomainName string, zerotierMembers map[string]zerotier.ControllerNetworkMember,
	aaaaRecords map[string][]string,
) map[string][]string {
	addressDomainNames := make(map[string][]string)
	for subname, records := range aaaaRecords {
		for _, ipAddress := range records {
			addressDomainNames[ipAddress] = append(addressDomainNames[ipAddress], subname)
		}
	}
	memberDomainNames := make(map[string][]string)
	for memberAddress, member := range zerotierMembers {
		for _, ipAddress := range *member.IpAssignments {
			for _, subname := range addressDomainNames[ipAddress] {
				domainName := subname + "." + zoneDomainName
				memberDomainNames[memberAddress] = append(memberDomainNames[memberAddress], domainName)
			}
		}
	}
	return memberDomainNames
}

func addNDPAddresses(
	id string, v6AssignMode zerotier.V6AssignMode,
	members map[string]zerotier.ControllerNetworkMember,
) error {
	n6plane := v6AssignMode.N6plane
	if n6plane != nil && *n6plane {
		for _, member := range members {
			n6PlaneAddress, err := zerotier.Get6Plane(id, *member.Address)
			if err != nil {
				return err
			}

			if member.IpAssignments == nil {
				ipAssignments := []string{n6PlaneAddress}
				member.IpAssignments = &ipAssignments
			} else {
				*member.IpAssignments = append(*member.IpAssignments, n6PlaneAddress)
			}
		}
	}
	return nil
}

type Member struct {
	ZerotierMember zerotier.ControllerNetworkMember
	DomainNames    []string
}

func getMemberRecords(
	ctx context.Context, zoneDomainName string, controller ztcontrollers.Controller,
	network zerotier.ControllerNetwork, memberAddresses []string,
	subnameRRsets map[string][]desec.RRset,
	c *ztc.Client,
) (map[string]Member, error) {
	zerotierMembers, err := c.GetNetworkMembers(ctx, controller, *network.Id, memberAddresses)
	if err != nil {
		return nil, err
	}
	if err = addNDPAddresses(*network.Id, *network.V6AssignMode, zerotierMembers); err != nil {
		return nil, err
	}

	aaaaRecords, err := getRecordsOfType(subnameRRsets, "AAAA")
	if err != nil {
		return nil, err
	}
	memberDomainNames := identifyMemberDomainNames(zoneDomainName, zerotierMembers, aaaaRecords)
	members := make(map[string]Member)
	for memberAddress, zerotierMember := range zerotierMembers {
		members[memberAddress] = Member{
			ZerotierMember: zerotierMember,
			DomainNames:    memberDomainNames[memberAddress],
		}
	}
	return members, nil
}

// Network DNS

func checkNamedByDNS(
	ctx context.Context, networkName, networkID string, c *desecc.Client,
) (bool, error) {
	domainSuffix := "." + c.Config.DomainName
	if !strings.HasSuffix(networkName, domainSuffix) {
		return false, nil
	}

	subname := strings.TrimSuffix(networkName, domainSuffix)
	txtRRset, err := c.GetRRset(ctx, subname, "TXT")
	if err != nil {
		return false, err
	}

	if txtRRset != nil {
		parsedID, txtHasNetworkID := client.GetNetworkID(txtRRset.Records)
		return txtHasNetworkID && (parsedID == networkID), nil
	}

	return false, nil
}

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
	namedByDNS, err := checkNamedByDNS(ctx, networkName, networkID, dc)
	if err != nil || !namedByDNS {
		return
	}
	networkDNS.Named = true

	txtRecords, err := getRecordsOfType(subnameRRsets, "TXT")
	if err != nil {
		return
	}
	confirmedSubname := strings.TrimSuffix(networkName, "."+dc.Config.DomainName)
	networkDNS.Aliases = identifyNetworkAliases(networkID, confirmedSubname, txtRecords)
	aliases := make(map[string]bool, len(networkDNS.Aliases))
	for _, alias := range networkDNS.Aliases {
		aliases[alias] = true
	}

	subdomains, err := client.GetSubdomains(ctx, subnameRRsets, dc, c, cc)
	if err != nil {
		return
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

	return
}

type NetworkData struct {
	Controller       ztcontrollers.Controller
	Network          zerotier.ControllerNetwork
	Members          map[string]Member
	JSONPrintedRules string
	DomainName       string
	NetworkDNS       NetworkDNS
}

func getNetworkData(
	ctx context.Context, address, id string,
	c *ztc.Client, cc *ztcontrollers.Client, dc *desecc.Client,
) (*NetworkData, error) {
	controller, err := cc.FindControllerByAddress(ctx, address)
	if err != nil {
		return nil, err
	}

	eg, egctx := errgroup.WithContext(ctx)
	var network *zerotier.ControllerNetwork
	var memberAddresses []string
	var subnameRRsets map[string][]desec.RRset
	eg.Go(func() (err error) {
		network, memberAddresses, err = c.GetNetworkInfo(ctx, *controller, id)
		return
	})
	eg.Go(func() (err error) {
		subnameRRsets, err = dc.GetRRsets(ctx)
		return
	})
	if err = eg.Wait(); err != nil {
		return nil, err
	}
	if network == nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "zerotier network not found")
	}
	rules, err := json.MarshalIndent(*network.Rules, "", "  ")
	if err != nil {
		return nil, err
	}

	eg, egctx = errgroup.WithContext(ctx)
	var members map[string]Member
	eg.Go(func() (err error) {
		members, err = getMemberRecords(
			ctx, dc.Config.DomainName, *controller, *network, memberAddresses, subnameRRsets, c,
		)
		return
	})
	var networkDNS NetworkDNS
	eg.Go(func() (err error) {
		networkDNS, err = getNetworkDNSRecords(
			egctx, *network.Id, *network.Name, subnameRRsets, c, cc, dc,
		)
		return
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return &NetworkData{
		Controller:       *controller,
		Network:          *network,
		Members:          members,
		JSONPrintedRules: string(rules),
		DomainName:       dc.Config.DomainName,
		NetworkDNS:       networkDNS,
	}, nil
}

func (h *Handlers) HandleNetworkGet() auth.AuthAwareHandler {
	t := "networks/network.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Parse params
		id := c.Param("id")
		address := ztc.GetControllerAddress(id)

		// Run queries
		networkData, err := getNetworkData(c.Request().Context(), address, id, h.ztc, h.ztcc, h.dc)
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, *networkData, a)
	}
}

func nameNetwork(
	ctx context.Context, controller ztcontrollers.Controller, id string, name string,
	c *ztc.Client, dc *desecc.Client,
) error {
	if len(name) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot remove name from network")
	}

	// Check to see if the network was already named by DNS
	fqdn := name + "." + dc.Config.DomainName
	txtRRset, err := dc.GetRRset(ctx, name, "TXT")
	if err != nil {
		return errors.Wrapf(
			err, "couldn't check cache for DNS TXT RRset at %s for network %s", fqdn, id,
		)
	}
	if txtRRset != nil {
		if _, hasID := client.GetNetworkID(txtRRset.Records); hasID {
			return echo.NewHTTPError(http.StatusBadRequest, "name is already used by another network")
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
		return errors.Wrapf(err, "couldn't create a DNS TXT RRset at %s for network %s", fqdn, id)
	}
	return c.UpdateNetwork(
		ctx, controller, id, zerotier.SetControllerNetworkJSONRequestBody{Name: &fqdn},
	)
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

			// Redirect user
			return c.Redirect(http.StatusSeeOther, "/networks")
		}
	}
}

func setNetworkRules(
	ctx context.Context, controller ztcontrollers.Controller,
	id string, jsonRules string, c *ztc.Client,
) error {
	rules := make([]map[string]interface{}, 0)
	if err := json.Unmarshal([]byte(jsonRules), &rules); err != nil {
		return err
	}
	err := c.UpdateNetwork(
		ctx, controller, id, zerotier.SetControllerNetworkJSONRequestBody{Rules: &rules},
	)
	if err != nil {
		return err
	}
	return nil
}

func (h *Handlers) HandleNetworkRulesPost() echo.HandlerFunc {
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
		if err = setNetworkRules(ctx, *controller, id, c.FormValue("rules"), h.ztc); err != nil {
			return err
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s#network-%s-rules", id, id))
	}
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
		if err = nameNetwork(ctx, *controller, id, c.FormValue("name"), h.ztc, h.dc); err != nil {
			return err
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s", id))
	}
}

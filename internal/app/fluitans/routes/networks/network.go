package networks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/netip"
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

// StringSet Utility

type StringSet map[string]struct{}

func NewStringSet(strings []string) StringSet {
	set := make(map[string]struct{})
	for _, s := range strings {
		set[s] = struct{}{}
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

func (ss StringSet) Difference(set StringSet) StringSet {
	difference := make(map[string]struct{})
	for s := range ss {
		if _, ok := set[s]; !ok {
			difference[s] = struct{}{}
		}
	}
	return difference
}

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

func identifyAddressDomainNames(
	subnameRRsets map[string][]desec.RRset,
) (addressDomainNames map[string][]string, err error) {
	aaaaRecords, err := getRecordsOfType(subnameRRsets, "AAAA")
	if err != nil {
		return nil, err
	}
	aRecords, err := getRecordsOfType(subnameRRsets, "A")
	if err != nil {
		return nil, err
	}

	addressDomainNames = make(map[string][]string)
	for subname, records := range aaaaRecords {
		for _, ipAddress := range records {
			addressDomainNames[ipAddress] = append(addressDomainNames[ipAddress], subname)
		}
	}
	for subname, records := range aRecords {
		for _, ipAddress := range records {
			addressDomainNames[ipAddress] = append(addressDomainNames[ipAddress], subname)
		}
	}
	return addressDomainNames, nil
}

func identifyDomainNames(
	zoneDomainName string, member zerotier.ControllerNetworkMember,
	addressDomainNames map[string][]string,
) (domainNames []string, subnames []string) {
	domainNames = make([]string, 0)
	subnames = make([]string, 0)
	domainNameAdded := make(map[string]struct{})
	for _, ipAddress := range *member.IpAssignments {
		for _, subname := range addressDomainNames[ipAddress] {
			domainName := subname + "." + zoneDomainName
			if _, alreadyAdded := domainNameAdded[domainName]; alreadyAdded {
				continue
			}
			domainNames = append(domainNames, domainName)
			subnames = append(subnames, subname)
			domainNameAdded[domainName] = struct{}{}
		}
	}
	return domainNames, subnames
}

func calculateNDPAddresses(
	networkID string, sixplane, rfc4193 bool, memberAddress string,
) (ndpAddresses []string, err error) {
	if !sixplane && !rfc4193 {
		return nil, nil
	}

	const ndpModes = 2
	ndpAddresses = make([]string, 0, ndpModes)
	if sixplane {
		sixplaneAddress, err := zerotier.Get6Plane(networkID, memberAddress)
		if err != nil {
			return nil, err
		}
		ndpAddresses = append(ndpAddresses, sixplaneAddress)
	}
	if rfc4193 {
		rfc4193Address, err := zerotier.GetRFC4193(networkID, memberAddress)
		if err != nil {
			return nil, err
		}
		ndpAddresses = append(ndpAddresses, rfc4193Address)
	}

	return ndpAddresses, nil
}

func calculateIPAddresses(
	networkID string, v6AssignMode zerotier.V6AssignMode, member zerotier.ControllerNetworkMember,
) (allIPAddresses []string, ndpAddresses []string, err error) {
	sixplane := (v6AssignMode.N6plane != nil) && *(v6AssignMode.N6plane)
	rfc4193 := (v6AssignMode.Rfc4193 != nil) && *(v6AssignMode.Rfc4193)
	ndpAddresses, err = calculateNDPAddresses(networkID, sixplane, rfc4193, *member.Address)
	if err != nil {
		return nil, nil, err
	}
	if member.IpAssignments == nil {
		return ndpAddresses, nil, nil
	}
	return append(ndpAddresses, *member.IpAssignments...), ndpAddresses, nil
}

type DNSUpdate struct {
	Type      string
	Operation string
	Record    string
}

func (u DNSUpdate) String() string {
	return fmt.Sprintf("%s: %s %s", u.Type, u.Operation, u.Record)
}

type Member struct {
	ZerotierMember zerotier.ControllerNetworkMember
	NDPAddresses   []string
	DomainNames    []string
	DNSUpdates     map[string][]DNSUpdate
}

func splitIPAddresses(rawAddresses []string) (ipv4 []string, ipv6 []string, err error) {
	ipv4 = make([]string, 0, len(rawAddresses))
	ipv6 = make([]string, 0, len(rawAddresses))
	for _, rawAddress := range rawAddresses {
		address, err := netip.ParseAddr(rawAddress)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "couldn't parse IP address %s", rawAddress)
		}
		if address.Is4() {
			ipv4 = append(ipv4, address.String())
			continue
		}
		if address.Is6() {
			ipv6 = append(ipv6, address.String())
			continue
		}
	}
	return ipv4, ipv6, nil
}

func planDNSUpdates(
	member zerotier.ControllerNetworkMember, subnames []string, domainNames []string,
	subnameRRsets map[string][]desec.RRset,
) (domainNameUpdates map[string][]DNSUpdate, err error) {
	ipv4Addresses, ipv6Addresses, err := splitIPAddresses(*member.IpAssignments)
	if err != nil {
		return nil, errors.Wrapf(
			err, "found invalid IP address for network member %s", *member.Address,
		)
	}
	aaaaExpected := NewStringSet(ipv6Addresses)
	aExpected := NewStringSet(ipv4Addresses)
	domainNameUpdates = make(map[string][]DNSUpdate)
	for i, subname := range subnames {
		var aaaaActual StringSet
		var aActual StringSet
		for _, rrset := range subnameRRsets[subname] {
			if rrset.Type == "AAAA" {
				aaaaActual = NewStringSet(rrset.Records)
			}
			if rrset.Type == "A" {
				aActual = NewStringSet(rrset.Records)
			}
		}
		domainName := domainNames[i]
		for address := range aaaaActual.Difference(aaaaExpected) {
			domainNameUpdates[domainName] = append(domainNameUpdates[domainName], DNSUpdate{
				Type:      "AAAA",
				Operation: "remove",
				Record:    address,
			})
		}
		for address := range aaaaExpected.Difference(aaaaActual) {
			domainNameUpdates[domainName] = append(domainNameUpdates[domainName], DNSUpdate{
				Type:      "AAAA",
				Operation: "add",
				Record:    address,
			})
		}
		for address := range aActual.Difference(aExpected) {
			domainNameUpdates[domainName] = append(domainNameUpdates[domainName], DNSUpdate{
				Type:      "A",
				Operation: "remove",
				Record:    address,
			})
		}
		for address := range aExpected.Difference(aActual) {
			domainNameUpdates[domainName] = append(domainNameUpdates[domainName], DNSUpdate{
				Type:      "A",
				Operation: "add",
				Record:    address,
			})
		}
	}
	return domainNameUpdates, nil
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
	addressDomainNames, err := identifyAddressDomainNames(subnameRRsets)
	if err != nil {
		return nil, err
	}

	memberNDPAddresses := make(map[string][]string)
	members := make(map[string]Member)
	for memberAddress, zerotierMember := range zerotierMembers {
		allIPAddresses, ndpAddresses, err := calculateIPAddresses(
			*network.Id, *network.V6AssignMode, zerotierMember,
		)
		if err != nil {
			return nil, err
		}
		zerotierMember.IpAssignments = &allIPAddresses
		// identifyDomainNames assumes the member's IP assignments include any assigned NDP addresses
		domainNames, subnames := identifyDomainNames(zoneDomainName, zerotierMember, addressDomainNames)
		dnsUpdates, err := planDNSUpdates(
			zerotierMember, subnames, domainNames, subnameRRsets,
		)
		if err != nil {
			return nil, errors.Wrapf(
				err, "couldn't calculate dns record updates needed for network %s member %s",
				*network.Id, memberAddress,
			)
		}
		members[memberAddress] = Member{
			ZerotierMember: zerotierMember,
			NDPAddresses:   ndpAddresses,
			DomainNames:    domainNames,
			DNSUpdates:     dnsUpdates,
		}
		memberNDPAddresses[memberAddress] = ndpAddresses
	}
	return members, nil
}

func SortNetworkMembers(members map[string]Member) (addresses []string, sorted []Member) {
	addresses = make([]string, 0, len(members))
	for address := range members {
		addresses = append(addresses, address)
	}
	sort.Slice(addresses, func(i, j int) bool {
		return client.CompareSubnamesAndAddresses(
			members[addresses[i]].DomainNames, addresses[i],
			members[addresses[j]].DomainNames, addresses[j],
		)
	})
	sorted = make([]Member, 0, len(addresses))
	for _, address := range addresses {
		sorted = append(sorted, members[address])
	}
	return addresses, sorted
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
		return NetworkDNS{}, err
	}
	networkDNS.Named = true

	txtRecords, err := getRecordsOfType(subnameRRsets, "TXT")
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
	Members          []Member
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
		network, memberAddresses, err = c.GetNetworkInfo(ctx, *controller, id)
		return err
	})
	eg.Go(func() (err error) {
		subnameRRsets, err = dc.GetRRsets(ctx)
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
		members, err := getMemberRecords(
			ctx, dc.Config.DomainName, *controller, *network, memberAddresses, subnameRRsets, c,
		)
		_, vd.Members = SortNetworkMembers(members)
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

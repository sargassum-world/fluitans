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

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	desecc "github.com/sargassum-eco/fluitans/internal/clients/desec"
	ztc "github.com/sargassum-eco/fluitans/internal/clients/zerotier"
	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/desec"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
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
			return nil, fmt.Errorf("unexpected number of RRsets for record")
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
				domainName := fmt.Sprintf("%s.%s", subname, zoneDomainName)
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
	domainSuffix := fmt.Sprintf(".%s", c.Config.DomainName)
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
	confirmedSubname := strings.TrimSuffix(networkName, fmt.Sprintf(".%s", dc.Config.DomainName))
	networkDNS.Aliases = identifyNetworkAliases(networkID, confirmedSubname, txtRecords)
	aliases := make(map[string]bool, len(networkDNS.Aliases))
	for _, alias := range networkDNS.Aliases {
		aliases[alias] = true
	}

	subdomains, err := client.GetSubdomains(ctx, subnameRRsets, dc, c, cc)
	if err != nil {
		return
	}
	networkSubname := strings.TrimSuffix(networkName, fmt.Sprintf(".%s", dc.Config.DomainName))
	networkDNS.DeviceSubdomains = make(map[string]client.Subdomain)
	for _, subdomain := range subdomains {
		if subdomain.Subname != networkSubname && !strings.HasSuffix(
			subdomain.Subname, fmt.Sprintf(".%s", networkSubname),
		) && !aliases[subdomain.Subname] {
			// Subdomain is unrelated to this network
			continue
		}

		if strings.HasSuffix(subdomain.Subname, fmt.Sprintf(".d.%s", networkSubname)) {
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

func (s *Service) getNetwork(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "networks/network.page.tmpl"
	te.Require(t)
	return func(c echo.Context) error {
		// Check authentication & authorization
		a, _, err := auth.GetWithSession(c, s.sc)
		if err != nil {
			return err
		}

		// Extract context
		ctx := c.Request().Context()

		// Parse params
		id := c.Param("id")
		address := ztc.GetControllerAddress(id)

		// Run queries
		networkData, err := getNetworkData(ctx, address, id, s.ztc, s.ztcc, s.dc)
		if err != nil {
			return err
		}

		// Produce output
		return route.Render(c, t, *networkData, a, te, g)
	}, nil
}

func nameNetwork(
	ctx context.Context, controller ztcontrollers.Controller, id string, name string,
	c *ztc.Client, dc *desecc.Client,
) error {
	if len(name) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "cannot remove name from network")
	}

	// Check to see if the network was already named by DNS
	fqdn := fmt.Sprintf("%s.%s", name, dc.Config.DomainName)
	txtRRset, err := dc.GetRRset(ctx, name, "TXT")
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(
			"couldn't check cache for DNS TXT RRset at %s for network %s", fqdn, id,
		))
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
		return errors.Wrap(err, fmt.Sprintf(
			"couldn't create a DNS TXT RRset at %s for network %s", fqdn, id,
		))
	}
	return c.UpdateNetwork(
		ctx, controller, id, zerotier.SetControllerNetworkJSONRequestBody{Name: &fqdn},
	)
}

func (s *Service) postNetwork(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	return func(c echo.Context) error {
		// Check authentication & authorization
		if err := auth.RequireAuthorized(c, s.sc); err != nil {
			return err
		}

		// Extract context
		ctx := c.Request().Context()

		// Parse params
		id := c.Param("id")
		address := ztc.GetControllerAddress(id)
		state := c.FormValue("state")

		// Run queries
		controller, err := s.ztcc.FindControllerByAddress(ctx, address)
		if err != nil {
			return err
		}

		switch state {
		default:
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"invalid network state %s", state,
			))
		case "deleted":
			if err = s.ztc.DeleteNetwork(ctx, *controller, id, s.ztcc); err != nil {
				// TODO: add a tombstone to the TXT RRset?
				return err
			}
			return c.Redirect(http.StatusSeeOther, "/networks")
		}
	}, nil
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

func (s *Service) postNetworkRules(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	return func(c echo.Context) error {
		// Check authentication & authorization
		if err := auth.RequireAuthorized(c, s.sc); err != nil {
			return err
		}

		// Extract context
		ctx := c.Request().Context()

		// Parse params
		id := c.Param("id")
		address := ztc.GetControllerAddress(id)

		// Run queries
		controller, err := s.ztcc.FindControllerByAddress(ctx, address)
		if err != nil {
			return err
		}
		if err = setNetworkRules(ctx, *controller, id, c.FormValue("rules"), s.ztc); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s#network-%s-rules", id, id))
	}, nil
}

func (s *Service) postNetworkName(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	return func(c echo.Context) error {
		// Check authentication & authorization
		if err := auth.RequireAuthorized(c, s.sc); err != nil {
			return err
		}

		// Extract context
		ctx := c.Request().Context()

		// Parse params
		id := c.Param("id")
		address := ztc.GetControllerAddress(id)

		// Run queries
		controller, err := s.ztcc.FindControllerByAddress(ctx, address)
		if err != nil {
			return err
		}
		if err = nameNetwork(
			ctx, *controller, id, c.FormValue("name"), s.ztc, s.dc,
		); err != nil {
			return err
		}

		return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/networks/%s", id))
	}, nil
}

package client

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"

	ztc "github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/desec"
	"github.com/sargassum-world/fluitans/pkg/zerotier"
)

// DNS Update

func determineExpectedRRsets(
	member zerotier.ControllerNetworkMember, subnames []string, ttl int,
) (expectedRRsets []desec.RRset, err error) {
	const approxRRsetsPerSubname = 2
	expectedRRsets = make([]desec.RRset, 0, approxRRsetsPerSubname*len(subnames))
	for _, subname := range subnames {
		rrsets, err := NewMemberNameRRsets(member, subname, ttl)
		if err != nil {
			return nil, errors.Wrapf(
				err, "couldn't make AAAA and A rrsets for network member %s", *member.Address,
			)
		}
		expectedRRsets = append(expectedRRsets, rrsets...)
	}
	return expectedRRsets, nil
}

type DNSUpdate struct {
	Type      string
	Operation string
	Record    string
}

func (u DNSUpdate) String() string {
	return fmt.Sprintf("%s: %s %s", u.Type, u.Operation, u.Record)
}

func planDNSUpdates(
	member zerotier.ControllerNetworkMember, subnames []string, domainNames []string,
	subnameRRsets map[string][]desec.RRset,
) (domainNameUpdates map[string][]DNSUpdate, err error) {
	ipv4Addresses, ipv6Addresses, err := SplitIPAddresses(*member.IpAssignments)
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

// Member

type Member struct {
	ZerotierMember zerotier.ControllerNetworkMember
	NDPAddresses   []string
	DomainNames    []string
	ExpectedRRsets []desec.RRset
	DNSUpdates     map[string][]DNSUpdate
}

func IdentifyAddressDomainNames(
	subnameRRsets map[string][]desec.RRset,
) (addressDomainNames map[string][]string, err error) {
	aaaaRecords, err := GetRecordsOfType(subnameRRsets, "AAAA")
	if err != nil {
		return nil, err
	}
	aRecords, err := GetRecordsOfType(subnameRRsets, "A")
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

func IdentifyDomainNames(
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

func GetMemberRecords(
	ctx context.Context, zoneDomainName string, controller ztcontrollers.Controller,
	network zerotier.ControllerNetwork, memberAddresses []string,
	subnameRRsets map[string][]desec.RRset,
	c *ztc.Client,
) (map[string]Member, error) {
	zerotierMembers, err := c.GetNetworkMembers(ctx, controller, *network.Id, memberAddresses)
	if err != nil {
		return nil, err
	}
	addressDomainNames, err := IdentifyAddressDomainNames(subnameRRsets)
	if err != nil {
		return nil, err
	}

	memberNDPAddresses := make(map[string][]string)
	members := make(map[string]Member)
	for memberAddress, zerotierMember := range zerotierMembers {
		allIPAddresses, ndpAddresses, err := ztc.CalculateIPAddresses(
			*network.Id, *network.V6AssignMode, zerotierMember,
		)
		if err != nil {
			return nil, err
		}
		zerotierMember.IpAssignments = &allIPAddresses
		domainNames, subnames := IdentifyDomainNames(zoneDomainName, zerotierMember, addressDomainNames)
		expectedRRsets, err := determineExpectedRRsets(
			zerotierMember, subnames, int(c.Config.DNS.DeviceTTL),
		)
		if err != nil {
			return nil, errors.Wrapf(
				err, "couldn't determine expected dns records for network %s member %s",
				*network.Id, memberAddress,
			)
		}
		dnsUpdates, err := planDNSUpdates(zerotierMember, subnames, domainNames, subnameRRsets)
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
			ExpectedRRsets: expectedRRsets,
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
		return CompareSubnamesAndAddresses(
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

func NewMemberNameRRsets(
	member zerotier.ControllerNetworkMember, memberSubname string,
	dnsTTL int,
) (rrsets []desec.RRset, err error) {
	ipv4Addresses, ipv6Addresses, err := SplitIPAddresses(*member.IpAssignments)
	if err != nil {
		return nil, errors.Wrapf(err, "found invalid IP address for network member %s", *member.Address)
	}

	return []desec.RRset{
		{
			Subname: memberSubname,
			Type:    "AAAA",
			Ttl:     &dnsTTL,
			Records: ipv6Addresses,
		},
		{
			Subname: memberSubname,
			Type:    "A",
			Ttl:     &dnsTTL,
			Records: ipv4Addresses,
		},
	}, nil
}

func NetworkNamedByDNS(
	networkID, networkName, domainName string, subnameRRsets map[string][]desec.RRset,
) bool {
	domainSuffix := "." + domainName
	if !strings.HasSuffix(networkName, domainSuffix) {
		return false
	}

	subname := strings.TrimSuffix(networkName, domainSuffix)
	rrsets, ok := subnameRRsets[subname]
	if !ok {
		return false
	}
	for _, rrset := range rrsets {
		if rrset.Type == "TXT" {
			parsedID, txtHasNetworkID := GetNetworkID(rrset.Records)
			return txtHasNetworkID && (parsedID == networkID)
		}
	}

	return false
}

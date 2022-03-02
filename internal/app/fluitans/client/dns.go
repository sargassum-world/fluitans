package client

import (
	"context"
	"sort"
	"strings"

	desecc "github.com/sargassum-world/fluitans/internal/clients/desec"
	ztc "github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/desec"
	"github.com/sargassum-world/fluitans/pkg/zerotier"
)

func GetReverseDomainNameFragments(domainName string) []string {
	fragments := strings.Split(domainName, ".")
	for i, j := 0, len(fragments)-1; i < j; i, j = i+1, j-1 {
		fragments[i], fragments[j] = fragments[j], fragments[i]
	}
	return fragments
}

func GetNetworkIDs(subnameRRsets map[string][]desec.RRset) map[string]string {
	networkIDs := make(map[string]string)
	for subname, rrsets := range subnameRRsets {
		for _, rrset := range rrsets {
			if rrset.Type == "TXT" {
				networkID, has := GetNetworkID(rrset.Records)
				if has {
					networkIDs[subname] = networkID
				}
				break // no need to check the non-TXT rrsets for this subname
			}
		}
	}
	return networkIDs
}

func SortSubnameRRsets(
	rrsets map[string][]desec.RRset, recordTypes []string,
) ([]string, [][]desec.RRset) {
	keys := make([]string, 0, len(rrsets))
	for key := range rrsets {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		a := GetReverseDomainNameFragments(keys[i])
		b := GetReverseDomainNameFragments(keys[j])
		k := 0
		for k = 0; k < len(a) && k < len(b); k++ {
			if a[k] < b[k] {
				return true
			}

			if a[k] > b[k] {
				return false
			}
		}
		return len(a) < len(b)
	})
	sorted := make([][]desec.RRset, 0, len(keys))
	for _, key := range keys {
		sorted = append(sorted, desecc.FilterAndSortRRsets(rrsets[key], recordTypes))
	}
	return keys, sorted
}

type Subdomain struct {
	Subname       string
	RRsets        []desec.RRset
	IsNetworkName bool
	Controller    *ztcontrollers.Controller
	Network       *zerotier.ControllerNetwork
}

func GetSubdomains(
	ctx context.Context, subnameRRsets map[string][]desec.RRset,
	c *desecc.Client, zc *ztc.Client, zcc *ztcontrollers.Client,
) ([]Subdomain, error) {
	ids := GetNetworkIDs(subnameRRsets)
	sortedKeys, sortedSubnameRRsets := SortSubnameRRsets(subnameRRsets, c.Cache.RecordTypes)
	networks, controllers, err := GetNetworks(ctx, ids, zc, zcc)
	if err != nil {
		return nil, err
	}

	subnames := make([]Subdomain, len(sortedKeys))
	for i, key := range sortedKeys {
		_, hasNetworkID := ids[key]
		subnames[i] = Subdomain{
			Subname:       key,
			RRsets:        sortedSubnameRRsets[i],
			IsNetworkName: hasNetworkID,
			Network:       networks[key],
			Controller:    controllers[key],
		}
	}
	return subnames, nil
}

package client

import (
	"context"

	"github.com/pkg/errors"

	desecc "github.com/sargassum-world/fluitans/internal/clients/desec"
	ztc "github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/desec"
	"github.com/sargassum-world/fluitans/pkg/zerotier"
)

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
	sortedKeys, sortedSubnameRRsets := desecc.SortSubnameRRsets(subnameRRsets, c.Cache.RecordTypes)
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

func GetRecordsOfType(
	subnameRRsets map[string][]desec.RRset, rrsetType string,
) (records map[string][]string, err error) {
	records = make(map[string][]string)
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

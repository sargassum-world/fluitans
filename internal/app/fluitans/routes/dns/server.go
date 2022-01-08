package dns

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	desecc "github.com/sargassum-eco/fluitans/internal/clients/desec"
	ztc "github.com/sargassum-eco/fluitans/internal/clients/zerotier"
	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/internal/models"
	"github.com/sargassum-eco/fluitans/pkg/desec"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/slidingwindows"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

type APILimiterStats struct {
	ReadLimiterFillRatios  []slidingwindows.KeyFillRatio
	ReadWaitSec            float64
	WriteLimiterFillRatios []slidingwindows.KeyFillRatio
	WriteBatchWaitSec      float64
}

type Subdomain struct {
	Subname       string
	RRsets        []desec.RRset
	IsNetworkName bool
	Controller    *ztcontrollers.Controller
	Network       *zerotier.ControllerNetwork
}

type ServerData struct {
	Server           models.DNSServer
	Domain           desec.Domain
	DesecAPISettings desecc.DesecAPISettings
	APILimiterStats  APILimiterStats
	ApexRRsets       []desec.RRset
	Subdomains       []Subdomain
}

func getReverseDomainNameFragments(domainName string) []string {
	fragments := strings.Split(domainName, ".")
	for i, j := 0, len(fragments)-1; i < j; i, j = i+1, j-1 {
		fragments[i], fragments[j] = fragments[j], fragments[i]
	}
	return fragments
}

func sortSubnameRRsets(
	rrsets map[string][]desec.RRset, recordTypes []string,
) ([]string, [][]desec.RRset) {
	keys := make([]string, 0, len(rrsets))
	for key := range rrsets {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		a := getReverseDomainNameFragments(keys[i])
		b := getReverseDomainNameFragments(keys[j])
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

func getNetworkIDs(subnameRRsets map[string][]desec.RRset) map[string]string {
	networkIDs := make(map[string]string)
	for subname, rrsets := range subnameRRsets {
		for _, rrset := range rrsets {
			if rrset.Type == "TXT" {
				networkID, has := client.GetNetworkID(rrset.Records)
				if has {
					networkIDs[subname] = networkID
				}
				break // no need to check the non-TXT rrsets for this subname
			}
		}
	}
	return networkIDs
}

func getNetworks(
	ctx context.Context, networkIDs map[string]string, c *ztc.Client, cc *ztcontrollers.Client,
) (map[string]*zerotier.ControllerNetwork, map[string]*ztcontrollers.Controller, error) {
	type IDAssociation struct {
		Subname string
		ID      string
	}
	idAssociations := make([]IDAssociation, 0, len(networkIDs))
	for subname, id := range networkIDs {
		idAssociations = append(idAssociations, IDAssociation{
			Subname: subname,
			ID:      id,
		})
	}

	eg, ctx := errgroup.WithContext(ctx)
	controllers := make([]*ztcontrollers.Controller, len(networkIDs))
	networks := make([]*zerotier.ControllerNetwork, len(networkIDs))
	for i, association := range idAssociations {
		eg.Go(func(i int, id string) func() error {
			return func() error {
				address := ztc.GetControllerAddress(id)
				controller, err := cc.FindControllerByAddress(ctx, address)
				if err != nil {
					return err
				}
				controllers[i] = controller

				network, err := c.GetNetwork(ctx, *controller, id)
				if err != nil {
					return err
				}
				networks[i] = network
				return nil
			}
		}(i, association.ID))
	}
	if err := eg.Wait(); err != nil {
		return nil, nil, err
	}

	// Transform the results into a more usable shape
	keyedNetworks := make(map[string]*zerotier.ControllerNetwork, len(networkIDs))
	for i, association := range idAssociations {
		keyedNetworks[association.Subname] = networks[i]
	}
	keyedControllers := make(map[string]*ztcontrollers.Controller, len(networkIDs))
	for i, association := range idAssociations {
		keyedControllers[association.Subname] = controllers[i]
	}
	return keyedNetworks, keyedControllers, nil
}

func getSubdomains(
	ctx context.Context, subnameRRsets map[string][]desec.RRset,
	c *desecc.Client, zc *ztc.Client, zcc *ztcontrollers.Client,
) ([]Subdomain, error) {
	ids := getNetworkIDs(subnameRRsets)
	sortedKeys, sortedSubnameRRsets := sortSubnameRRsets(subnameRRsets, c.Cache.RecordTypes)
	networks, controllers, err := getNetworks(ctx, ids, zc, zcc)
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

func getServerData(
	ctx context.Context, c *desecc.Client, zc *ztc.Client, zcc *ztcontrollers.Client,
) (*ServerData, error) {
	readLimiter := c.ReadLimiter
	writeLimiter := c.WriteLimiter
	desecDomain, err := c.GetDomain(ctx)
	if err != nil {
		return nil, err
	}

	subnameRRsets, err := c.GetRRsets(ctx)
	if err != nil {
		return nil, err
	}
	apexRRsets := desecc.FilterAndSortRRsets(subnameRRsets[""], c.Cache.RecordTypes)
	delete(subnameRRsets, "")

	subdomains, err := getSubdomains(ctx, subnameRRsets, c, zc, zcc)
	if err != nil {
		return nil, err
	}

	return &ServerData{
		Server:           c.Config.DNSServer,
		Domain:           *desecDomain,
		DesecAPISettings: c.Config.APISettings,
		APILimiterStats: APILimiterStats{
			ReadLimiterFillRatios:  readLimiter.EstimateFillRatios(time.Now()),
			ReadWaitSec:            readLimiter.EstimateWaitDuration(time.Now(), 1).Seconds(),
			WriteLimiterFillRatios: writeLimiter.EstimateFillRatios(time.Now()),
			WriteBatchWaitSec: desecc.CalculateBatchWaitDuration(
				writeLimiter, c.Config.APISettings.WriteSoftQuota,
			).Seconds(),
		},
		ApexRRsets: apexRRsets,
		Subdomains: subdomains,
	}, nil
}

func getServer(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "dns/server.page.tmpl"
	err := te.RequireSegments("dns.getServer", t)
	if err != nil {
		return nil, err
	}

	switch app := g.App.(type) {
	default:
		return nil, client.NewUnexpectedGlobalsTypeError(app)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()

			// Run queries
			serverData, err := getServerData(
				ctx, app.Clients.Desec, app.Clients.Zerotier, app.Clients.ZTControllers,
			)
			if err != nil {
				return err
			}

			// Produce output
			return route.Render(c, t, *serverData, te, g)
		}, nil
	}
}

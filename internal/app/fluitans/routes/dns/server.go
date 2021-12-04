package dns

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	desecc "github.com/sargassum-eco/fluitans/internal/clients/desec"
	"github.com/sargassum-eco/fluitans/internal/models"
	"github.com/sargassum-eco/fluitans/pkg/desec"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/slidingwindows"
)

type APILimiterStats struct {
	ReadLimiterFillRatios  []slidingwindows.KeyFillRatio
	ReadWaitSec            float64
	WriteLimiterFillRatios []slidingwindows.KeyFillRatio
	WriteBatchWaitSec      float64
}

type ServerData struct {
	Server           models.DNSServer
	Domain           desec.Domain
	DesecAPISettings desecc.DesecAPISettings
	APILimiterStats  APILimiterStats
	ApexRRsets       []desec.RRset
	SubnameRRsets    [][]desec.RRset
}

func getReverseDomainNameFragments(domainName string) []string {
	fragments := strings.Split(domainName, ".")
	for i, j := 0, len(fragments)-1; i < j; i, j = i+1, j-1 {
		fragments[i], fragments[j] = fragments[j], fragments[i]
	}
	return fragments
}

func sortSubnameRRsets(rrsets map[string][]desec.RRset, recordTypes []string) [][]desec.RRset {
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
	return sorted
}

func getServerData(ctx context.Context, c *desecc.Client) (*ServerData, error) {
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
	sortedSubnameRRsets := sortSubnameRRsets(subnameRRsets, c.Cache.RecordTypes)

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
		ApexRRsets:    apexRRsets,
		SubnameRRsets: sortedSubnameRRsets,
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
			serverData, err := getServerData(ctx, app.Clients.Desec)
			if err != nil {
				return err
			}

			// Produce output
			return route.Render(c, t, *serverData, te, g)
		}, nil
	}
}

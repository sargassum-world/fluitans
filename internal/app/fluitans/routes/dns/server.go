package dns

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/caching"
	"github.com/sargassum-eco/fluitans/internal/fingerprint"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
	"github.com/sargassum-eco/fluitans/pkg/desec"
	"github.com/sargassum-eco/fluitans/pkg/slidingwindows"
)

type APILimiterStats struct {
	ReadLimiterFillRatios  []slidingwindows.KeyFillRatio
	ReadWaitSec            float64
	WriteLimiterFillRatios []slidingwindows.KeyFillRatio
	WriteBatchWaitSec      float64
}

type ServerData struct {
	Server           client.DNSServer
	Domain           desec.Domain
	DesecAPISettings client.DesecAPISettings
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

func sortSubnameRRsets(rrsets map[string][]desec.RRset) [][]desec.RRset {
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
		sorted = append(sorted, client.FilterAndSortRRsets(rrsets[key]))
	}
	return sorted
}

func getServerData(
	ctx context.Context, cg *client.Globals, l echo.Logger,
) (*ServerData, error) {
	readLimiter := cg.RateLimiters[client.DesecReadLimiterName]
	writeLimiter := cg.RateLimiters[client.DesecWriteLimiterName]
	domain, err := client.NewDNSDomain(
		cg.RateLimiters[client.DesecReadLimiterName], cg.Cache,
	)
	if err != nil {
		return nil, err
	}

	desecDomain, err := client.GetDomain(ctx, *domain, l)
	if err != nil {
		return nil, err
	}

	subnameRRsets, err := client.GetRRsets(ctx, *domain, l)
	if err != nil {
		return nil, err
	}
	apexRRsets := client.FilterAndSortRRsets(subnameRRsets[""])
	delete(subnameRRsets, "")
	sortedSubnameRRsets := sortSubnameRRsets(subnameRRsets)

	return &ServerData{
		Server:           domain.Server,
		Domain:           *desecDomain,
		DesecAPISettings: domain.APISettings,
		APILimiterStats: APILimiterStats{
			ReadLimiterFillRatios:  readLimiter.EstimateFillRatios(time.Now()),
			ReadWaitSec:            readLimiter.EstimateWaitDuration(time.Now(), 1).Seconds(),
			WriteLimiterFillRatios: writeLimiter.EstimateFillRatios(time.Now()),
			WriteBatchWaitSec: client.CalculateBatchWaitDuration(
				writeLimiter, domain.APISettings.WriteSoftQuota,
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
	tte, ok := te[t]
	if !ok {
		return nil, errors.Wrap(
			te.NewNotFoundError(t), "couldn't find template for dns.getServer",
		)
	}

	switch app := g.App.(type) {
	default:
		return nil, errors.Errorf("app globals are of unexpected type %T", g.App)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()
			l := c.Logger()

			// Run queries
			serverData, err := getServerData(ctx, app, l)
			if err != nil {
				return err
			}

			// Handle Etag
			etagData, err := json.Marshal(serverData)
			if err != nil {
				return err
			}

			if noContent, err := caching.ProcessEtag(c, tte, fingerprint.Compute(etagData)); noContent {
				return err
			}

			// Render template
			return c.Render(http.StatusOK, t, struct {
				Meta   template.Meta
				Embeds template.Embeds
				Data   ServerData
			}{
				Meta: template.Meta{
					Path:       c.Request().URL.Path,
					DomainName: client.GetEnvVarDomainName(),
				},
				Embeds: g.Embeds,
				Data:   *serverData,
			})
		}, nil
	}
}

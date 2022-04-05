package dns

import (
	"context"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/client"
	desecc "github.com/sargassum-world/fluitans/internal/clients/desec"
	ztc "github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/internal/models"
	"github.com/sargassum-world/fluitans/pkg/desec"
	"github.com/sargassum-world/fluitans/pkg/slidingwindows"
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
	Subdomains       []client.Subdomain
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

	subdomains, err := client.GetSubdomains(ctx, subnameRRsets, c, zc, zcc)
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

func (h *Handlers) HandleServerGet() auth.HTTPHandlerFunc {
	t := "dns/server.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Run queries
		serverData, err := getServerData(c.Request().Context(), h.dc, h.ztc, h.ztcc)
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, *serverData, a)
	}
}

package dns

import (
	"context"
	"reflect"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest/handling"
	"github.com/sargassum-world/godest/turbostreams"

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

type ServerViewData struct {
	Server           models.DNSServer
	Domain           desec.Domain
	DesecAPISettings desecc.DesecAPISettings
	APILimiterStats  APILimiterStats
	ApexRRsets       []desec.RRset
	Subdomains       []client.Subdomain
}

func getAPILimiterStats(c *desecc.Client) APILimiterStats {
	readLimiter := c.ReadLimiter
	writeLimiter := c.WriteLimiter
	return APILimiterStats{
		ReadLimiterFillRatios:  readLimiter.EstimateFillRatios(time.Now()),
		ReadWaitSec:            readLimiter.EstimateWaitDuration(time.Now(), 1).Seconds(),
		WriteLimiterFillRatios: writeLimiter.EstimateFillRatios(time.Now()),
		WriteBatchWaitSec: desecc.CalculateBatchWaitDuration(
			writeLimiter, c.Config.APISettings.WriteSoftQuota,
		).Seconds(),
	}
}

func getServerViewData(
	ctx context.Context, c *desecc.Client, zc *ztc.Client, zcc *ztcontrollers.Client,
) (vd ServerViewData, err error) {
	vd.Server = c.Config.DNSServer

	desecDomain, err := c.GetDomain(ctx)
	if err != nil {
		return ServerViewData{}, err
	}
	if desecDomain == nil {
		return ServerViewData{}, errors.New("couldn't get desec domain")
	}
	vd.Domain = *desecDomain

	vd.DesecAPISettings = c.Config.APISettings
	vd.APILimiterStats = getAPILimiterStats(c)

	subnameRRsets, err := c.GetRRsets(ctx)
	if err != nil {
		return ServerViewData{}, err
	}
	vd.ApexRRsets = desecc.FilterAndSortRRsets(subnameRRsets[""], c.Cache.RecordTypes)

	delete(subnameRRsets, "")
	if vd.Subdomains, err = client.GetSubdomains(ctx, subnameRRsets, c, zc, zcc); err != nil {
		return ServerViewData{}, err
	}

	return vd, nil
}

func (h *Handlers) HandleServerGet() auth.HTTPHandlerFunc {
	t := "dns/server.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Run queries
		serverView, err := getServerViewData(c.Request().Context(), h.dc, h.ztc, h.ztcc)
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, serverView, a)
	}
}

const (
	readQuotasPartial  = "dns/desec-read-quotas.partial.tmpl"
	writeQuotasPartial = "dns/desec-write-quotas.partial.tmpl"
)

func replaceServerInfoStream(c *desecc.Client) []turbostreams.Message {
	stats := getAPILimiterStats(c)
	return []turbostreams.Message{
		{
			Action:   turbostreams.ActionReplace,
			Target:   "/dns/server/info/read-quotas",
			Template: readQuotasPartial,
			Data: map[string]interface{}{
				"DesecAPISettings": c.Config.APISettings,
				"APILimiterStats":  stats,
			},
		},
		{
			Action:   turbostreams.ActionReplace,
			Target:   "/dns/server/info/write-quotas",
			Template: writeQuotasPartial,
			Data: map[string]interface{}{
				"DesecAPISettings": c.Config.APISettings,
				"APILimiterStats":  stats,
			},
		},
	}
}

func (h *Handlers) HandleServerInfoPub() turbostreams.HandlerFunc {
	h.r.MustHave(readQuotasPartial, writeQuotasPartial)
	return func(c turbostreams.Context) error {
		// Make change trackers
		var prevStats APILimiterStats

		// Publish periodically
		const pubInterval = 1 * time.Second
		return handling.Repeat(c.Context(), pubInterval, func() (done bool, err error) {
			// Check for changes
			stats := getAPILimiterStats(h.dc)
			if reflect.DeepEqual(prevStats, stats) {
				return false, nil
			}
			prevStats = stats

			// Publish changes
			messages := replaceServerInfoStream(h.dc)
			c.Publish(messages...)
			return false, nil
		})
	}
}

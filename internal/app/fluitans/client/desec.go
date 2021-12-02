package client

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/desec"
	"github.com/sargassum-eco/fluitans/pkg/slidingwindows"
)

// Rate-Limiting

var (
	DesecReadLimiterName  = "DesecRead"
	DesecWriteLimiterName = "DesecWrite"
)

func CalculateBatchWaitDuration(
	ml *slidingwindows.MultiLimiter, batchMinFillRatio float32,
) time.Duration {
	rates := ml.Rates()
	keyFillRatios := ml.EstimateFillRatios(time.Now())
	var maxWaitDuration time.Duration = 0
	for _, kf := range keyFillRatios {
		if kf.FillRatio > float64(batchMinFillRatio) {
			rate := rates[kf.Key]
			maxBatches := float64(rate.AmountPerTime) * (1.0 - float64(batchMinFillRatio))
			waitDuration := time.Duration(float64(rate.Time) / maxBatches)
			if waitDuration > maxWaitDuration {
				maxWaitDuration = waitDuration
			}
		}
	}
	return maxWaitDuration
}

func newReadRateLimitError(retryWaitSec float64) error {
	return echo.NewHTTPError(
		http.StatusTooManyRequests,
		fmt.Sprintf(
			"too many read requests have been issued to the DeSEC API. Try again in %f sec.",
			retryWaitSec,
		),
	)
}

func getRetryWait(c echo.Context, header http.Header) float64 {
	retryWait := header.Get("Retry-After")
	floatWidth := 64
	retryWaitSec, err := strconv.ParseFloat(retryWait, floatWidth)
	if err != nil {
		c.Logger().Error(errors.Wrap(err, "couldn't get parse Retry-After header value from deSEC API"))
		retryWaitSec = 1.0 // default to a retry period of 1 sec
	}
	return retryWaitSec
}

// DNSDomain

type DesecAPISettings struct {
	ReadCacheTTL   time.Duration
	WriteSoftQuota float32
}

type DNSDomain struct {
	Server      DNSServer
	APISettings DesecAPISettings
	DomainName  string
	Cache       *Cache
	ReadLimiter *slidingwindows.MultiLimiter
}

func NewDNSDomain(
	readLimiter *slidingwindows.MultiLimiter, cache *Cache,
) (*DNSDomain, error) {
	server, err := GetEnvVarDNSServer()
	if err != nil {
		return nil, err
	}

	apiSettings, err := GetEnvVarDesecAPISettings()
	if err != nil {
		return nil, err
	}

	domainName := GetEnvVarDomainName()
	domain := DNSDomain{
		Server:      *server,
		APISettings: *apiSettings,
		DomainName:  domainName,
		Cache:       cache,
		ReadLimiter: readLimiter,
	}
	return &domain, nil
}

func (domain *DNSDomain) makeClientWithResponses() (*desec.ClientWithResponses, error) {
	return desec.NewAuthClientWithResponses(
		domain.Server.Server, domain.Server.Authtoken,
	)
}

func (domain *DNSDomain) handleDesecMissingDomainError(res http.Response) error {
	if res.StatusCode == http.StatusNotFound {
		domain.Cache.SetNonexistentDomainByName(
			domain.DomainName, domain.Server.NetworkCostWeight,
			domain.APISettings.ReadCacheTTL,
		)
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf(
			"couldn't find domain %s", domain.DomainName,
		))
	}

	return nil
}

func (domain *DNSDomain) handleDesecMissingRRsetError(
	res http.Response, subname, recordType string,
) error {
	if res.StatusCode == http.StatusNotFound {
		domain.Cache.SetNonexistentDomainByName(
			domain.DomainName, domain.Server.NetworkCostWeight,
			domain.APISettings.ReadCacheTTL,
		)
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf(
			"couldn't find %s RRset for %s.%s", recordType, subname, domain.DomainName,
		))
	}

	return nil
}

func (domain *DNSDomain) handleDesecClientError(c echo.Context, res http.Response) error {
	switch res.StatusCode {
	case http.StatusTooManyRequests:
		retryWaitSec := getRetryWait(c, res.Header)
		// The read limiter expected not to be throttled, so its estimates of API usage
		// need to be adjusted upwards
		domain.ReadLimiter.Throttled(time.Now(), retryWaitSec)
		return newReadRateLimitError(retryWaitSec)
	case http.StatusBadRequest:
		// TODO: handle pagination when there's a Link: header
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	return nil
}

func (domain *DNSDomain) tryAddLimitedRead() error {
	maybeAllowed := domain.ReadLimiter.MaybeAllowed(time.Now(), 1)
	if !maybeAllowed || !domain.ReadLimiter.TryAdd(time.Now(), 1) {
		return newReadRateLimitError(
			domain.ReadLimiter.EstimateWaitDuration(time.Now(), 1).Seconds(),
		)
	}

	return nil
}

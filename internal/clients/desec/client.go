// Package desec provides a high-level client to the deSEC API
package desec

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/desec"
	"github.com/sargassum-eco/fluitans/pkg/framework/clientcache"
	"github.com/sargassum-eco/fluitans/pkg/framework/log"
	"github.com/sargassum-eco/fluitans/pkg/slidingwindows"
)

type Client struct {
	Config       Config
	Logger       log.Logger
	Cache        *Cache
	ReadLimiter  *slidingwindows.MultiLimiter
	WriteLimiter *slidingwindows.MultiLimiter
}

func (c *Client) handleDesecMissingDomainError(res http.Response) error {
	if res.StatusCode == http.StatusNotFound {
		c.Cache.SetNonexistentDomainByName(c.Config.DomainName)
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf(
			"couldn't find domain %s", c.Config.DomainName,
		))
	}

	return nil
}

func (c *Client) handleDesecMissingRRsetError(
	res http.Response, subname, recordType string,
) error {
	if res.StatusCode == http.StatusNotFound {
		c.Cache.SetNonexistentDomainByName(c.Config.DomainName)
		return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf(
			"couldn't find %s RRset for %s.%s", recordType, subname, c.Config.DomainName,
		))
	}

	return nil
}

func (c *Client) handleDesecClientError(res http.Response, l log.Logger) error {
	switch res.StatusCode {
	case http.StatusTooManyRequests:
		retryWaitSec := getRetryWait(res.Header, l)
		// The read limiter expected not to be throttled, so its estimates of API usage
		// need to be adjusted upwards
		c.ReadLimiter.Throttled(time.Now(), retryWaitSec)
		return newReadRateLimitError(retryWaitSec)
	case http.StatusBadRequest:
		// TODO: handle pagination when there's a Link: header
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	return nil
}

func (c *Client) tryAddLimitedRead() error {
	maybeAllowed := c.ReadLimiter.MaybeAllowed(time.Now(), 1)
	if !maybeAllowed || !c.ReadLimiter.TryAdd(time.Now(), 1) {
		return newReadRateLimitError(
			c.ReadLimiter.EstimateWaitDuration(time.Now(), 1).Seconds(),
		)
	}

	return nil
}

func MakeClient(domainName string, cache clientcache.Cache, l log.Logger) (*Client, error) {
	config, err := GetConfig(domainName)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up desec client config")
	}

	clientCache := Cache{
		Cache:       cache,
		CostWeight:  config.DNSServer.NetworkCostWeight,
		TTL:         config.APISettings.ReadCacheTTL,
		RecordTypes: config.RecordTypes,
	}
	readLimiter := desec.NewReadLimiter(0)
	return &Client{
		Config:       *config,
		Logger:       l,
		Cache:        &clientCache,
		ReadLimiter:  readLimiter,
		WriteLimiter: desec.NewRRSetWriteLimiter(0),
	}, nil
}

// Rate-Limiting

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

func getRetryWait(header http.Header, l log.Logger) float64 {
	retryWait := header.Get("Retry-After")
	floatWidth := 64
	retryWaitSec, err := strconv.ParseFloat(retryWait, floatWidth)
	if err != nil {
		l.Error(errors.Wrap(err, "couldn't get parse Retry-After header value from deSEC API"))
		retryWaitSec = 1.0 // default to a retry period of 1 sec
	}
	return retryWaitSec
}

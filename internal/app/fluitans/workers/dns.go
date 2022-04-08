package workers

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/handling"
	"github.com/sargassum-world/fluitans/internal/clients/desec"
)

func PrefetchDNSRecords(ctx context.Context, c *desec.Client) error {
	const retryInterval = 5 * time.Second
	return handling.Repeat(ctx, retryInterval, func() (done bool, err error) {
		if _, err := c.GetRRsets(ctx); err != nil {
			c.Logger.Error(errors.Wrap(err, "couldn't prefetch DNS records for cache"))
			return false, nil
		}

		return true, nil
	})
}

func TestWriteLimiter(ctx context.Context, c *desec.Client) error {
	const writeInterval = 5 * time.Second
	writeLimiter := c.WriteLimiter
	return handling.Repeat(ctx, writeInterval, func() (done bool, err error) {
		_ = writeLimiter.TryAdd(time.Now(), 1)
		/*if writeLimiter.TryAdd(time.Now(), 1) {
			// fmt.Printf("Bumped the write limiter: %+v\n", writeLimiter.EstimateFillRatios(time.Now()))
		} else {
			fmt.Printf(
				"Write limiter throttled: wait %f sec\n",
				writeLimiter.EstimateWaitDuration(time.Now(), 1).Seconds(),
			)
		}*/
		return false, nil
	})
}

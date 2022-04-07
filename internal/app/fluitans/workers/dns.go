package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/internal/clients/desec"
)

func PrefetchDNSRecords(ctx context.Context, c *desec.Client) error {
	const retryInterval = 5000
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		if _, err := c.GetRRsets(ctx); err != nil {
			c.Logger.Error(errors.Wrap(err, "couldn't prefetch DNS records for cache"))
			time.Sleep(retryInterval * time.Millisecond)
			continue
		}

		return nil
	}
}

func TestWriteLimiter(ctx context.Context, c *desec.Client) error {
	const writeInterval = 5000
	writeLimiter := c.WriteLimiter
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		if writeLimiter.TryAdd(time.Now(), 1) {
			// fmt.Printf("Bumped the write limiter: %+v\n", writeLimiter.EstimateFillRatios(time.Now()))
		} else {
			fmt.Printf(
				"Write limiter throttled: wait %f sec\n",
				writeLimiter.EstimateWaitDuration(time.Now(), 1).Seconds(),
			)
		}
		time.Sleep(writeInterval * time.Millisecond)
	}
}

package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/internal/clients/desec"
)

func PrefetchDNSRecords(c *desec.Client) error {
	const retryInterval = 5000
	for {
		if _, err := c.GetRRsets(context.Background()); err != nil {
			c.Logger.Error(errors.Wrap(err, "couldn't prefetch DNS records for cache"))
			time.Sleep(retryInterval * time.Millisecond)
			continue
		}

		break
	}
	return nil
}

func TestWriteLimiter(c *desec.Client) error {
	const writeInterval = 5000
	writeLimiter := c.WriteLimiter
	for {
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

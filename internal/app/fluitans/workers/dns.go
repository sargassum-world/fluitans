package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/clients/desec"
)

func PrefetchDNSRecords(c *desec.Client) {
	for {
		if _, err := c.GetRRsets(context.Background()); err != nil {
			c.Logger.Error(errors.Wrap(err, "couldn't prefetch DNS records for cache"))
			continue
		}

		break
	}
}

func TestWriteLimiter(c *desec.Client) {
	var writeInterval time.Duration = 5000
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

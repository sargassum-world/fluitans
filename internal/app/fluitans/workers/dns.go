package workers

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
)

func PrefetchDNSRecords(cg *client.Globals) {
	l := cg.Logger
	for {
		if _, err := client.GetRRsets(context.Background(), cg.DNSDomain, l); err != nil {
			l.Error(errors.Wrap(err, "couldn't prefetch DNS records for cache"))
			continue
		}

		break
	}
}

func TestWriteLimiter(cg *client.Globals) {
	var writeInterval time.Duration = 5000
	writeLimiter := cg.RateLimiters[client.DesecWriteLimiterName]
	for {
		if writeLimiter.TryAdd(time.Now(), 1) {
			/*fmt.Printf(
				"Bumped the write limiter: %+v\n",
				writeLimiter.EstimateFillRatios(time.Now()),
			)*/
		} else {
			fmt.Printf(
				"Write limiter throttled: wait %f sec\n",
				writeLimiter.EstimateWaitDuration(time.Now(), 1).Seconds(),
			)
		}
		time.Sleep(writeInterval * time.Millisecond)
	}
}

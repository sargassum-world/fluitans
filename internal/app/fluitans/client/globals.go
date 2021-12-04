// Package client contains client code for external APIs
package client

import (
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/conf"
	"github.com/sargassum-eco/fluitans/pkg/slidingwindows"
)

type Globals struct {
	Config       conf.Config
	Cache        *Cache
	RateLimiters map[string]*slidingwindows.MultiLimiter
	DNSDomain    *DNSDomain
}

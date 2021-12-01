// Package client contains client code for external APIs
package client

import (
	"github.com/sargassum-eco/fluitans/pkg/slidingwindows"
)

type Globals struct {
	Cache        *Cache
	RateLimiters map[string]*slidingwindows.MultiLimiter
}

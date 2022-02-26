// Package clientcache provides support for caching requests as a client
package clientcache

import (
	"time"
)

type Cache interface {
	SetEntry(key string, value interface{}, costWeight float32, ttl time.Duration) error
	UnsetEntry(key string)
	SetNonexistentEntry(key string, cost float32, ttl time.Duration)
	GetEntry(key string, value interface{}) (bool, bool, error)
}

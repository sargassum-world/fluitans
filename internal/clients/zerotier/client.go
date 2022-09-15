// Package zerotier provides a high-level client for the Zerotier network controller API
package zerotier

import (
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/clientcache"
)

type Client struct {
	Config Config
	Logger godest.Logger
	Cache  *Cache
}

func NewClient(c Config, cache clientcache.Cache, l godest.Logger) *Client {
	return &Client{
		Config: c,
		Logger: l,
		Cache: &Cache{
			Cache: cache,
		},
	}
}

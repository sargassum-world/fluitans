// Package ztcontrollers provides a high-level client for management of Zerotier
// network controllers
package ztcontrollers

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

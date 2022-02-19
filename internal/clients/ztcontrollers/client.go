// Package ztcontrollers provides a high-level client for management of Zerotier
// network controllers
package ztcontrollers

import (
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework/clientcache"
	"github.com/sargassum-eco/fluitans/pkg/framework/log"
)

type Client struct {
	Config Config
	Logger log.Logger
	Cache  *Cache
}

func NewClient(cache clientcache.Cache, l log.Logger) (*Client, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up zerotier controllers client config")
	}

	return &Client{
		Config: *config,
		Logger: l,
		Cache: &Cache{
			Cache: cache,
		},
	}, nil
}

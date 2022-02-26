// Package ztcontrollers provides a high-level client for management of Zerotier
// network controllers
package ztcontrollers

import (
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework"
	"github.com/sargassum-eco/fluitans/pkg/framework/clientcache"
)

type Client struct {
	Config Config
	Logger framework.Logger
	Cache  *Cache
}

func NewClient(cache clientcache.Cache, l framework.Logger) (*Client, error) {
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

// Package zerotier provides a high-level client for the Zerotier network controller API
package zerotier

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

func MakeClient(cache clientcache.Cache, l log.Logger) (*Client, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up zerotier client config")
	}

	return &Client{
		Config: *config,
		Logger: l,
		Cache: &Cache{
			Cache: cache,
		},
	}, nil
}

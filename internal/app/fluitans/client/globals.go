// Package client contains client code for external APIs
package client

import (
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/conf"
	"github.com/sargassum-eco/fluitans/internal/clients/desec"
	"github.com/sargassum-eco/fluitans/internal/clients/zerotier"
	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/framework/clientcache"
	"github.com/sargassum-eco/fluitans/pkg/framework/log"
)

type Clients struct {
	Desec         *desec.Client
	Zerotier      *zerotier.Client
	ZTControllers *ztcontrollers.Client
}

type Globals struct {
	Config  conf.Config
	Clients *Clients
}

func NewGlobals(l log.Logger) (*Globals, error) {
	config, err := conf.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up application config")
	}

	cache, err := clientcache.NewRistrettoCache(config.Cache)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up client cache")
	}

	desecClient, err := desec.MakeClient(config.DomainName, cache, l)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up desec client")
	}

	ztClient, err := zerotier.MakeClient(cache, l)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up zerotier client")
	}

	ztcClient, err := ztcontrollers.MakeClient(cache, l)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up zerotier controllers client")
	}

	return &Globals{
		Config: *config,
		Clients: &Clients{
			Desec:         desecClient,
			Zerotier:      ztClient,
			ZTControllers: ztcClient,
		},
	}, nil
}

func NewUnexpectedGlobalsTypeError(g interface{}) error {
	return errors.Errorf("globals are of unexpected type %T", g)
}

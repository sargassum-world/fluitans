// Package client contains client code for external APIs
package client

import (
	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/conf"
	"github.com/sargassum-world/fluitans/internal/clients/authn"
	"github.com/sargassum-world/fluitans/internal/clients/desec"
	"github.com/sargassum-world/fluitans/internal/clients/sessions"
	"github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/clientcache"
)

type Clients struct {
	Authn         *authn.Client
	Desec         *desec.Client
	Sessions      *sessions.Client
	Zerotier      *zerotier.Client
	ZTControllers *ztcontrollers.Client
}

type Globals struct {
	Config  conf.Config
	Clients *Clients
}

func NewGlobals(l godest.Logger) (*Globals, error) {
	config, err := conf.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up application config")
	}

	cache, err := clientcache.NewRistrettoCache(config.Cache)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up client cache")
	}

	authnClient, err := authn.NewClient(l)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up authn client")
	}

	desecClient, err := desec.NewClient(config.DomainName, cache, l)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up desec client")
	}

	sessionsClient, err := sessions.NewMemStoreClient(l)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up sessions client")
	}

	ztClient, err := zerotier.NewClient(cache, l)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up zerotier client")
	}

	ztcClient, err := ztcontrollers.NewClient(cache, l)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up zerotier controllers client")
	}

	return &Globals{
		Config: config,
		Clients: &Clients{
			Authn:         authnClient,
			Desec:         desecClient,
			Sessions:      sessionsClient,
			Zerotier:      ztClient,
			ZTControllers: ztcClient,
		},
	}, nil
}
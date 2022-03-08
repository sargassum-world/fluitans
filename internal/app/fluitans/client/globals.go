// Package client contains client code for external APIs
package client

import (
	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/conf"
	"github.com/sargassum-world/fluitans/internal/clients/desec"
	"github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/authn"
	"github.com/sargassum-world/fluitans/pkg/godest/clientcache"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
)

type Clients struct {
	Authn         *authn.Client
	Desec         *desec.Client
	Sessions      *session.Client
	Zerotier      *zerotier.Client
	ZTControllers *ztcontrollers.Client
}

type Globals struct {
	Config  conf.Config
	Cache   clientcache.Cache
	Clients *Clients
}

func NewGlobals(l godest.Logger) (g *Globals, err error) {
	g = &Globals{}
	if g.Config, err = conf.GetConfig(); err != nil {
		err = errors.Wrap(err, "couldn't set up application config")
		return
	}
	if g.Cache, err = clientcache.NewRistrettoCache(g.Config.Cache); err != nil {
		err = errors.Wrap(err, "couldn't set up client cache")
		return
	}
	g.Clients = &Clients{}

	authnConfig, err := authn.GetConfig()
	if err != nil {
		err = errors.Wrap(err, "couldn't set up authn config")
		return
	}
	g.Clients.Authn = authn.NewClient(authnConfig)

	desecConfig, err := desec.GetConfig(g.Config.DomainName)
	if err != nil {
		err = errors.Wrap(err, "couldn't set up desec config")
		return
	}
	g.Clients.Desec = desec.NewClient(desecConfig, g.Cache, l)

	sessionsConfig, err := session.GetConfig()
	if err != nil {
		err = errors.Wrap(err, "couldn't set up sessions config")
		return
	}
	g.Clients.Sessions = session.NewMemStoreClient(sessionsConfig)

	ztConfig, err := zerotier.GetConfig()
	if err != nil {
		err = errors.Wrap(err, "couldn't set up zerotier config")
		return
	}
	g.Clients.Zerotier = zerotier.NewClient(ztConfig, g.Cache, l)

	ztcConfig, err := ztcontrollers.GetConfig()
	if err != nil {
		err = errors.Wrap(err, "couldn't set up zerotier controllers config")
		return
	}
	g.Clients.ZTControllers = ztcontrollers.NewClient(ztcConfig, g.Cache, l)

	return
}

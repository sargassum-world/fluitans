// Package client contains client code for external APIs
package client

import (
	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/conf"
	"github.com/sargassum-world/fluitans/internal/clients/desec"
	"github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
	"github.com/sargassum-world/fluitans/pkg/godest/authn"
	"github.com/sargassum-world/fluitans/pkg/godest/clientcache"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"
)

type Clients struct {
	Authn    *authn.Client
	Sessions *session.Client

	Desec         *desec.Client
	Zerotier      *zerotier.Client
	ZTControllers *ztcontrollers.Client
}

type Globals struct {
	Config       conf.Config
	Cache        clientcache.Cache
	ACCancellers *actioncable.Cancellers
	TSBroker     *turbostreams.Broker
	Clients      *Clients
}

func NewGlobals(l godest.Logger) (g *Globals, err error) {
	g = &Globals{}
	if g.Config, err = conf.GetConfig(); err != nil {
		return nil, errors.Wrap(err, "couldn't set up application config")
	}
	if g.Cache, err = clientcache.NewRistrettoCache(g.Config.Cache); err != nil {
		return nil, errors.Wrap(err, "couldn't set up client cache")
	}
	g.ACCancellers = actioncable.NewCancellers()
	g.TSBroker = turbostreams.NewBroker(l)
	g.Clients = &Clients{}

	authnConfig, err := authn.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up authn config")
	}
	g.Clients.Authn = authn.NewClient(authnConfig)

	sessionsConfig, err := session.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up sessions config")
	}
	g.Clients.Sessions = session.NewMemStoreClient(sessionsConfig)

	desecConfig, err := desec.GetConfig(g.Config.DomainName)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up desec config")
	}
	g.Clients.Desec = desec.NewClient(desecConfig, g.Cache, l)

	ztConfig, err := zerotier.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up zerotier config")
	}
	g.Clients.Zerotier = zerotier.NewClient(ztConfig, g.Cache, l)

	ztcConfig, err := ztcontrollers.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up zerotier controllers config")
	}
	g.Clients.ZTControllers = ztcontrollers.NewClient(ztcConfig, g.Cache, l)

	return g, nil
}

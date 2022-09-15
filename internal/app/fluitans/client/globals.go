// Package client contains client code for external APIs
package client

import (
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/actioncable"
	"github.com/sargassum-world/godest/authn"
	"github.com/sargassum-world/godest/clientcache"
	"github.com/sargassum-world/godest/session"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/conf"
	"github.com/sargassum-world/fluitans/internal/clients/desec"
	"github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
)

type Globals struct {
	Config conf.Config
	Cache  clientcache.Cache

	Sessions    session.Store
	CSRFChecker *session.CSRFTokenChecker
	Authn       *authn.Client

	ACCancellers *actioncable.Cancellers
	TSSigner     turbostreams.Signer
	TSBroker     *turbostreams.Broker

	Desec         *desec.Client
	Zerotier      *zerotier.Client
	ZTControllers *ztcontrollers.Client

	Logger godest.Logger
}

func NewGlobals(l godest.Logger) (g *Globals, err error) {
	g = &Globals{}
	if g.Config, err = conf.GetConfig(); err != nil {
		return nil, errors.Wrap(err, "couldn't set up application config")
	}
	if g.Cache, err = clientcache.NewRistrettoCache(g.Config.Cache); err != nil {
		return nil, errors.Wrap(err, "couldn't set up client cache")
	}

	sessionsConfig, err := session.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up sessions config")
	}
	g.Sessions = session.NewMemStore(sessionsConfig)
	g.CSRFChecker = session.NewCSRFTokenChecker(sessionsConfig)
	authnConfig, err := authn.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up authn config")
	}
	g.Authn = authn.NewClient(authnConfig)

	g.ACCancellers = actioncable.NewCancellers()
	tssConfig, err := turbostreams.GetSignerConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up turbo streams signer config")
	}
	g.TSSigner = turbostreams.NewSigner(tssConfig)
	g.TSBroker = turbostreams.NewBroker(l)

	desecConfig, err := desec.GetConfig(g.Config.DomainName)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up desec config")
	}
	g.Desec = desec.NewClient(desecConfig, g.Cache, l)
	ztConfig, err := zerotier.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up zerotier config")
	}
	g.Zerotier = zerotier.NewClient(ztConfig, g.Cache, l)
	ztcConfig, err := ztcontrollers.GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up zerotier controllers config")
	}
	g.ZTControllers = ztcontrollers.NewClient(ztcConfig, g.Cache, l)

	g.Logger = l
	return g, nil
}

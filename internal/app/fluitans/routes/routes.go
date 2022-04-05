// Package routes contains the route handlers for the Fluitans server.
package routes

import (
	"github.com/sargassum-world/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/assets"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/auth"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/cable"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/controllers"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/dns"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/home"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/networks"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"
)

type Handlers struct {
	r       godest.TemplateRenderer
	acc     *actioncable.Cancellers
	tsb     *turbostreams.Broker
	clients *client.Clients
	logger  godest.Logger
}

func New(
	r godest.TemplateRenderer, acc *actioncable.Cancellers, tsb *turbostreams.Broker,
	clients *client.Clients, logger godest.Logger,
) *Handlers {
	return &Handlers{
		r:       r,
		acc:     acc,
		tsb:     tsb,
		clients: clients,
		logger:  logger,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, tsr turbostreams.Router, em godest.Embeds) {
	assets.RegisterStatic(er, em)
	assets.NewTemplated(h.r).Register(er)
	cable.New(h.r, h.acc, h.tsb, h.clients.Sessions, h.logger).Register(er)
	home.New(h.r).Register(er, h.clients.Sessions)
	auth.New(h.r, h.acc, h.clients.Authn, h.clients.Sessions).Register(er)
	controllers.New(h.r, h.clients.ZTControllers, h.clients.Zerotier).Register(er, h.clients.Sessions)
	networks.New(
		h.r, h.clients.Desec, h.clients.Zerotier, h.clients.ZTControllers,
	).Register(er, tsr, h.clients.Sessions)
	dns.New(
		h.r, h.clients.Desec, h.clients.Zerotier, h.clients.ZTControllers,
	).Register(er, h.clients.Sessions)
}

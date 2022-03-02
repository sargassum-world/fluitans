// Package routes contains the route handlers for the Fluitans server.
package routes

import (
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/assets"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/auth"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/controllers"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/dns"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/home"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/networks"
	"github.com/sargassum-eco/fluitans/pkg/godest"
)

type Handlers struct {
	r       godest.TemplateRenderer
	clients *client.Clients
}

func New(r godest.TemplateRenderer, clients *client.Clients) *Handlers {
	return &Handlers{
		r:       r,
		clients: clients,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, em godest.Embeds) {
	assets.RegisterStatic(er, em)
	assets.NewTemplated(h.r).Register(er)
	home.New(h.r, h.clients.Sessions).Register(er)
	auth.New(h.r, h.clients.Authn, h.clients.Sessions).Register(er)
	controllers.New(h.r, h.clients.ZTControllers, h.clients.Zerotier, h.clients.Sessions).Register(er)
	networks.New(
		h.r, h.clients.Desec, h.clients.Zerotier, h.clients.ZTControllers, h.clients.Sessions,
	).Register(er)
	dns.New(
		h.r, h.clients.Desec, h.clients.Zerotier, h.clients.ZTControllers, h.clients.Sessions,
	).Register(er)
}

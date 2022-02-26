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

type Service struct {
	r       godest.TemplateRenderer
	clients *client.Clients
}

func NewService(r godest.TemplateRenderer, clients *client.Clients) *Service {
	return &Service{
		r:       r,
		clients: clients,
	}
}

func (s *Service) Register(er godest.EchoRouter, em godest.Embeds) {
	assets.RegisterStatic(er, em)
	assets.NewTemplatedService(s.r).Register(er)
	home.NewService(s.r, s.clients.Sessions).Register(er)
	auth.NewService(s.r, s.clients.Authn, s.clients.Sessions).Register(er)
	controllers.NewService(
		s.r, s.clients.ZTControllers, s.clients.Zerotier, s.clients.Sessions,
	).Register(er)
	networks.NewService(
		s.r, s.clients.Desec, s.clients.Zerotier, s.clients.ZTControllers, s.clients.Sessions,
	).Register(er)
	dns.NewService(
		s.r, s.clients.Desec, s.clients.Zerotier, s.clients.ZTControllers, s.clients.Sessions,
	).Register(er)
}

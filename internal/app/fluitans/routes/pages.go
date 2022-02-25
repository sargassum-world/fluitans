// Package routes contains the route handlers for the Fluitans server.
package routes

import (
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/auth"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/controllers"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/dns"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/home"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes/networks"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

type Service struct {
	clients *client.Clients
}

func NewService(clients *client.Clients) *Service {
	return &Service{
		clients: clients,
	}
}

func (s Service) Routes() []route.Templated {
	return route.CollectTemplated(
		home.NewService(s.clients.Sessions).Routes(),
		auth.NewService(s.clients.Authn, s.clients.Sessions).Routes(),
		controllers.NewService(
			s.clients.ZTControllers, s.clients.Zerotier, s.clients.Sessions,
		).Routes(),
		networks.NewService(
			s.clients.Desec, s.clients.Zerotier, s.clients.ZTControllers, s.clients.Sessions,
		).Routes(),
		dns.NewService(
			s.clients.Desec, s.clients.Zerotier, s.clients.ZTControllers, s.clients.Sessions,
		).Routes(),
	)
}

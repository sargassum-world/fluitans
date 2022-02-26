// Package auth contains the route handlers related to authentication and authorization.
package auth

import (
	"github.com/sargassum-eco/fluitans/internal/clients/authn"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/pkg/framework"
)

type Service struct {
	r  framework.TemplateRenderer
	ac *authn.Client
	sc *sessions.Client
}

func NewService(r framework.TemplateRenderer, ac *authn.Client, sc *sessions.Client) *Service {
	return &Service{
		r:  r,
		ac: ac,
		sc: sc,
	}
}

func (s *Service) Register(er framework.EchoRouter) {
	er.GET("/csrf", s.getCSRF())
	er.GET("/login", s.getLogin())
	er.POST("/sessions", s.postSessions())
}

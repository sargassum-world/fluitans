// Package auth contains the route handlers related to authentication and authorization.
package auth

import (
	"net/http"

	"github.com/sargassum-eco/fluitans/internal/clients/authn"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

type Service struct {
	ac *authn.Client
	sc *sessions.Client
}

func NewService(ac *authn.Client, sc *sessions.Client) *Service {
	return &Service{
		ac: ac,
		sc: sc,
	}
}

func (s *Service) Routes() []route.Templated {
	return []route.Templated{
		{
			Path:         "/csrf",
			Method:       http.MethodGet,
			HandlerMaker: s.getCSRF,
		},
		{
			Path:         "/login",
			Method:       http.MethodGet,
			HandlerMaker: s.getLogin,
			Templates:    []string{"auth/login.page.tmpl"},
		},
		{
			Path:         "/sessions",
			Method:       http.MethodPost,
			HandlerMaker: s.postSessions,
		},
	}
}

// Package auth contains the route handlers related to authentication and authorization.
package auth

import (
	"github.com/sargassum-eco/fluitans/internal/clients/authn"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/pkg/godest"
)

type Handlers struct {
	r  godest.TemplateRenderer
	ac *authn.Client
	sc *sessions.Client
}

func New(r godest.TemplateRenderer, ac *authn.Client, sc *sessions.Client) *Handlers {
	return &Handlers{
		r:  r,
		ac: ac,
		sc: sc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	er.GET("/csrf", h.HandleCSRFGet())
	er.GET("/login", h.HandleLoginGet())
	er.POST("/sessions", h.HandleSessionsPost())
}

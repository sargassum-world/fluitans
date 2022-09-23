// Package auth contains the route handlers related to authentication and authorization.
package auth

import (
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/actioncable"
	"github.com/sargassum-world/godest/authn"
	"github.com/sargassum-world/godest/session"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
)

type Handlers struct {
	r godest.TemplateRenderer

	ss *session.Store

	acc *actioncable.Cancellers
	ac  *authn.Client
}

func New(
	r godest.TemplateRenderer, ss *session.Store, acc *actioncable.Cancellers, ac *authn.Client,
) *Handlers {
	return &Handlers{
		r:   r,
		ss:  ss,
		acc: acc,
		ac:  ac,
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	er.GET("/csrf", h.HandleCSRFGet())
	er.GET("/login", auth.HandleHTTPWithSession(h.HandleLoginGet(), h.ss))
	er.POST("/sessions", h.HandleSessionsPost())
}

// Package auth contains the route handlers related to authentication and authorization.
package auth

import (
	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
	"github.com/sargassum-world/fluitans/pkg/godest/authn"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
)

type Handlers struct {
	r   godest.TemplateRenderer
	acc *actioncable.Cancellers
	ac  *authn.Client
	sc  *session.Client
}

func New(
	r godest.TemplateRenderer, acc *actioncable.Cancellers, ac *authn.Client, sc *session.Client,
) *Handlers {
	return &Handlers{
		r:   r,
		acc: acc,
		ac:  ac,
		sc:  sc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	er.GET("/csrf", h.HandleCSRFGet())
	er.GET("/login", auth.HandleHTTPWithSession(h.HandleLoginGet(), h.sc))
	er.POST("/sessions", h.HandleSessionsPost())
}

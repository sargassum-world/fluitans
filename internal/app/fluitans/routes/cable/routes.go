// Package cable contains the route handlers serving Action Cables over WebSockets
// by implementing the Action Cable Protocol (https://docs.anycable.io/misc/action_cable_protocol)
// on the server.
package cable

import (
	"github.com/gorilla/websocket"
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/actioncable"
	"github.com/sargassum-world/godest/session"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
)

type Handlers struct {
	r godest.TemplateRenderer

	ss *session.Store
	cc *session.CSRFTokenChecker

	acc *actioncable.Cancellers
	acs actioncable.Signer
	tsb *turbostreams.Broker

	wsu websocket.Upgrader

	l godest.Logger
}

func New(
	r godest.TemplateRenderer, ss *session.Store, cc *session.CSRFTokenChecker,
	acc *actioncable.Cancellers, acs actioncable.Signer, tsb *turbostreams.Broker, l godest.Logger,
) *Handlers {
	return &Handlers{
		r:   r,
		ss:  ss,
		cc:  cc,
		acc: acc,
		acs: acs,
		tsb: tsb,
		wsu: websocket.Upgrader{
			Subprotocols: actioncable.SupportedSubprotocols(),
			// TODO: add parameters to the upgrader as needed
		},
		l: l,
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	er.GET("/cable", auth.HandleHTTPWithSession(h.HandleCableGet(), h.ss))
}

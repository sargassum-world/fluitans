// Package cable contains the route handlers serving Action Cables over WebSockets
// by implementing the Action Cable Protocol (https://docs.anycable.io/misc/action_cable_protocol)
// on the server.
package cable

import (
	"github.com/gorilla/websocket"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"
)

type Handlers struct {
	r   godest.TemplateRenderer
	acc *actioncable.Cancellers
	tsb *turbostreams.Broker
	ss  session.Store
	l   godest.Logger
	wsu websocket.Upgrader
}

func New(
	r godest.TemplateRenderer, acc *actioncable.Cancellers, tsb *turbostreams.Broker,
	ss session.Store, l godest.Logger,
) *Handlers {
	return &Handlers{
		r:   r,
		acc: acc,
		tsb: tsb,
		ss:  ss,
		l:   l,
		wsu: websocket.Upgrader{
			Subprotocols: actioncable.Subprotocols(),
			// TODO: add parameters to the upgrader as needed
		},
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	er.GET("/cable", auth.HandleHTTPWithSession(h.HandleCableGet(), h.ss))
}

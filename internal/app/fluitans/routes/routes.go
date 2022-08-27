// Package routes contains the route handlers for the Fluitans server.
package routes

import (
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/assets"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/auth"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/cable"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/controllers"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/dns"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/home"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/routes/networks"
)

type Handlers struct {
	r       godest.TemplateRenderer
	globals *client.Globals
}

func New(r godest.TemplateRenderer, globals *client.Globals) *Handlers {
	return &Handlers{
		r:       r,
		globals: globals,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, tsr turbostreams.Router, em godest.Embeds) {
	acc := h.globals.ACCancellers
	ss := h.globals.Sessions
	ztcc := h.globals.ZTControllers
	ztc := h.globals.Zerotier
	dc := h.globals.Desec

	assets.RegisterStatic(er, em)
	assets.NewTemplated(h.r).Register(er)
	cable.New(
		h.r, ss, h.globals.CSRFChecker, acc, h.globals.TSSigner, h.globals.TSBroker, h.globals.Logger,
	).Register(er)
	home.New(h.r).Register(er, ss)
	auth.New(h.r, ss, acc, h.globals.Authn).Register(er)
	controllers.New(h.r, ztcc, ztc).Register(er, ss)
	networks.New(h.r, h.globals.TSBroker.Hub(), dc, ztc, ztcc).Register(er, tsr, ss)
	dns.New(h.r, dc, ztc, ztcc).Register(er, tsr, ss)

	tsr.UNSUB("/*", turbostreams.EmptyHandler)
}

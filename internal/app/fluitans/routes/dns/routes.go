// Package dns contains the route handlers related to DNS records
package dns

import (
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/session"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/handling"
	"github.com/sargassum-world/fluitans/internal/clients/desec"
	"github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
)

type Handlers struct {
	r godest.TemplateRenderer

	dc   *desec.Client
	ztc  *zerotier.Client
	ztcc *ztcontrollers.Client
}

func New(
	r godest.TemplateRenderer,
	dc *desec.Client, ztc *zerotier.Client, ztcc *ztcontrollers.Client,
) *Handlers {
	return &Handlers{
		r:    r,
		dc:   dc,
		ztc:  ztc,
		ztcc: ztcc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, tsr turbostreams.Router, ss *session.Store) {
	hr := auth.NewHTTPRouter(er, ss)
	haz := auth.RequireHTTPAuthz(ss)
	tsaz := auth.RequireTSAuthz(ss)
	hr.GET("/dns", h.HandleServerGet(), haz)
	tsr.SUB("/dns/server/info", turbostreams.EmptyHandler, tsaz)
	tsr.PUB("/dns/server/info", h.HandleServerInfoPub())
	tsr.MSG("/dns/server/info", handling.HandleTSMsg(h.r, ss), tsaz)
	er.POST("/dns/:subname/:type", h.HandleRRsetPost(), haz)
}

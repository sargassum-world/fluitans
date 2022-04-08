// Package dns contains the route handlers related to DNS records
package dns

import (
	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/internal/app/fluitans/handling"
	"github.com/sargassum-world/fluitans/internal/clients/desec"
	"github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"
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

func (h *Handlers) Register(er godest.EchoRouter, tsr turbostreams.Router, ss session.Store) {
	ar := auth.NewHTTPRouter(er, ss)
	az := auth.RequireHTTPAuthz(ss)
	atsz := auth.RequireTSAuthz(ss)
	ar.GET("/dns", h.HandleServerGet(), az)
	tsr.SUB("/dns/server/info", turbostreams.EmptyHandler, atsz)
	tsr.PUB("/dns/server/info", h.HandleServerInfoPub())
	tsr.MSG("/dns/server/info", handling.HandleTSMsg(h.r, ss), atsz)
	er.POST("/dns/:subname/:type", h.HandleRRsetPost(), az)
}

// Package dns contains the route handlers related to DNS records
package dns

import (
	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/internal/clients/desec"
	"github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
)

type Handlers struct {
	r    godest.TemplateRenderer
	dc   *desec.Client
	ztc  *zerotier.Client
	ztcc *ztcontrollers.Client
	sc   *session.Client
}

func New(
	r godest.TemplateRenderer,
	dc *desec.Client, ztc *zerotier.Client, ztcc *ztcontrollers.Client, sc *session.Client,
) *Handlers {
	return &Handlers{
		r:    r,
		dc:   dc,
		ztc:  ztc,
		ztcc: ztcc,
		sc:   sc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	ar := auth.NewRouter(er, h.sc)
	ar.GET("/dns", h.HandleServerGet(), auth.RequireAuthz(h.sc))
	er.POST("/dns/:subname/:type", h.HandleRRsetPost(), auth.RequireAuthz(h.sc))
}

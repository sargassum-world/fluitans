// Package dns contains the route handlers related to DNS records
package dns

import (
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/clients/desec"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/internal/clients/zerotier"
	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/godest"
)

type Handlers struct {
	r    godest.TemplateRenderer
	dc   *desec.Client
	ztc  *zerotier.Client
	ztcc *ztcontrollers.Client
	sc   *sessions.Client
}

func New(
	r godest.TemplateRenderer,
	dc *desec.Client, ztc *zerotier.Client, ztcc *ztcontrollers.Client, sc *sessions.Client,
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
	ar := auth.NewAuthAwareRouter(er, h.sc)
	ar.GET("/dns", h.HandleServerGet(), auth.RequireAuthz(h.sc))
	er.POST("/dns/:subname/:type", h.HandleRRsetPost(), auth.RequireAuthz(h.sc))
}

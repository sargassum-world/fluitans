// Package networks contains the route handlers related to ZeroTier networks.
package networks

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

	tsh *turbostreams.MessagesHub

	dc   *desec.Client
	ztc  *zerotier.Client
	ztcc *ztcontrollers.Client
}

func New(
	r godest.TemplateRenderer, tsh *turbostreams.MessagesHub,
	dc *desec.Client, ztc *zerotier.Client, ztcc *ztcontrollers.Client,
) *Handlers {
	return &Handlers{
		r:    r,
		tsh:  tsh,
		dc:   dc,
		ztc:  ztc,
		ztcc: ztcc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, tsr turbostreams.Router, ss *session.Store) {
	hr := auth.NewHTTPRouter(er, ss)
	haz := auth.RequireHTTPAuthz(ss)
	tsaz := auth.RequireTSAuthz(ss)
	hr.GET("/networks", h.HandleNetworksGet())
	er.POST("/networks", h.HandleNetworksPost(), haz)
	hr.GET("/networks/:id", h.HandleNetworkGet())
	er.POST("/networks/:id", h.HandleNetworkPost(), haz)
	er.POST("/networks/:id/name", h.HandleNetworkNamePost(), haz)
	hr.POST("/networks/:id/rules", h.HandleNetworkRulesPost(), haz)
	hr.POST("/networks/:id/devices", h.HandleDevicesPost(), haz)
	tsr.SUB("/networks/:id/devices", h.HandleDevicesSub(), tsaz)
	tsr.PUB("/networks/:id/devices", h.HandleDevicesPub())
	tsr.MSG("/networks/:id/devices", handling.HandleTSMsg(h.r, ss), tsaz)
	hr.POST("/networks/:id/devices/:address/authorization", h.HandleDeviceAuthorizationPost(), haz)
	hr.POST("/networks/:id/devices/:address/name", h.HandleDeviceNamePost(), haz)
}

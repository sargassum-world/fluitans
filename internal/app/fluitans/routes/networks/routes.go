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

	tsh *turbostreams.Hub

	dc   *desec.Client
	ztc  *zerotier.Client
	ztcc *ztcontrollers.Client
}

func New(
	r godest.TemplateRenderer, tsh *turbostreams.Hub,
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
	hr.POST("/networks/:id/routes", h.HandleNetworkRoutesPost(), haz)
	hr.POST("/networks/:id/autoip/v6-modes", h.HandleNetworkAutoIPv6ModesPost(), haz)
	hr.POST("/networks/:id/autoip/v4-modes", h.HandleNetworkAutoIPv4ModesPost(), haz)
	hr.POST("/networks/:id/autoip/pools", h.HandleNetworkAutoIPPoolsPost(), haz)
	hr.POST("/networks/:id/rules", h.HandleNetworkRulesPost(), haz)
	hr.POST("/networks/:id/devices", h.HandleDevicesPost(), haz)
	tsr.SUB("/networks/:id/devices", h.HandleDevicesSub(), tsaz)
	tsr.PUB("/networks/:id/devices", h.HandleDevicesPub())
	tsr.MSG("/networks/:id/devices", handling.HandleTSMsg(h.r, ss), tsaz)
	tsr.SUB("/networks/:id/devices/:address", h.HandleDeviceSub(), tsaz)
	tsr.PUB("/networks/:id/devices/:address", h.HandleDevicePub())
	tsr.MSG("/networks/:id/devices/:address", handling.HandleTSMsg(h.r, ss), tsaz)
	hr.POST("/networks/:id/devices/:address/authorization", h.HandleDeviceAuthorizationPost(), haz)
	hr.POST("/networks/:id/devices/:address/name", h.HandleDeviceNamePost(), haz)
}

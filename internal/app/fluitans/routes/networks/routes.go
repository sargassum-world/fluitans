// Package networks contains the route handlers related to ZeroTier networks.
package networks

import (
	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/internal/clients/desec"
	"github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"
)

type Handlers struct {
	r    godest.TemplateRenderer
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

func (h *Handlers) Register(er godest.EchoRouter, tsr turbostreams.Router, sc *session.Client) {
	ahr := auth.NewHTTPRouter(er, sc)
	atsr := auth.NewTSRouter(tsr, sc)
	ahz := auth.RequireHTTPAuthz(sc)
	atsz := auth.RequireTSAuthz(sc)
	ahr.GET("/networks", h.HandleNetworksGet())
	er.POST("/networks", h.HandleNetworksPost(), ahz)
	ahr.GET("/networks/:id", h.HandleNetworkGet())
	er.POST("/networks/:id", h.HandleNetworkPost(), ahz)
	er.POST("/networks/:id/name", h.HandleNetworkNamePost(), ahz)
	ahr.POST("/networks/:id/rules", h.HandleNetworkRulesPost(), ahz)
	ahr.POST("/networks/:id/devices", h.HandleDevicesPost(), ahz)
	atsr.SUB("/networks/:id/devices", h.HandleDevicesSub(), atsz)
	atsr.MSG("/networks/:id/devices", h.HandleDevicesMsg(), atsz)
	tsr.PUB("/networks/:id/devices", h.HandleDevicesPub())
	ahr.POST("/networks/:id/devices/:address/authorization", h.HandleDeviceAuthorizationPost(), ahz)
	ahr.POST("/networks/:id/devices/:address/name", h.HandleDeviceNamePost(), ahz)
}

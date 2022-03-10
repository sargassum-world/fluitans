// Package networks contains the route handlers related to ZeroTier networks.
package networks

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

func (h *Handlers) Register(er godest.EchoRouter, sc *session.Client) {
	ar := auth.NewRouter(er, sc)
	az := auth.RequireAuthz(sc)
	ar.GET("/networks", h.HandleNetworksGet())
	er.POST("/networks", h.HandleNetworksPost(), az)
	ar.GET("/networks/:id", h.HandleNetworkGet())
	er.POST("/networks/:id", h.HandleNetworkPost(), az)
	er.POST("/networks/:id/name", h.HandleNetworkNamePost(), az)
	er.POST("/networks/:id/rules", h.HandleNetworkRulesPost(), az)
	er.POST("/networks/:id/devices", h.HandleDevicesPost(), az)
	er.POST("/networks/:id/devices/:address/authorization", h.HandleDeviceAuthorizationPost(), az)
	er.POST("/networks/:id/devices/:address/name", h.HandleDeviceNamePost(), az)
}

// Package networks contains the route handlers related to ZeroTier networks.
package networks

import (
	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/internal/clients/desec"
	"github.com/sargassum-world/fluitans/internal/clients/sessions"
	"github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/godest"
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
	ar.GET("/networks", h.HandleNetworksGet())
	er.POST("/networks", h.HandleNetworksPost(), auth.RequireAuthz(h.sc))
	ar.GET("/networks/:id", h.HandleNetworkGet())
	er.POST("/networks/:id", h.HandleNetworkPost(), auth.RequireAuthz(h.sc))
	er.POST("/networks/:id/name", h.HandleNetworkNamePost(), auth.RequireAuthz(h.sc))
	er.POST("/networks/:id/rules", h.HandleNetworkRulesPost(), auth.RequireAuthz(h.sc))
	er.POST("/networks/:id/devices", h.HandleDevicesPost(), auth.RequireAuthz(h.sc))
	er.POST(
		"/networks/:id/devices/:address/authorization", h.HandleDeviceAuthorizationPost(),
		auth.RequireAuthz(h.sc),
	)
	er.POST("/networks/:id/devices/:address/name", h.HandleDeviceNamePost(), auth.RequireAuthz(h.sc))
}

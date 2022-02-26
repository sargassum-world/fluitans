// Package networks contains the route handlers related to ZeroTier networks.
package networks

import (
	"github.com/sargassum-eco/fluitans/internal/clients/desec"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/internal/clients/zerotier"
	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/godest"
)

type Service struct {
	r    godest.TemplateRenderer
	dc   *desec.Client
	ztc  *zerotier.Client
	ztcc *ztcontrollers.Client
	sc   *sessions.Client
}

func NewService(
	r godest.TemplateRenderer,
	dc *desec.Client, ztc *zerotier.Client, ztcc *ztcontrollers.Client, sc *sessions.Client,
) *Service {
	return &Service{
		r:    r,
		dc:   dc,
		ztc:  ztc,
		ztcc: ztcc,
		sc:   sc,
	}
}

func (s *Service) Register(er godest.EchoRouter) {
	er.GET("/networks", s.getNetworks())
	er.POST("/networks", s.postNetworks())
	er.GET("/networks/:id", s.getNetwork())
	er.POST("/networks/:id", s.postNetwork())
	er.POST("/networks/:id/name", s.postNetworkName())
	er.POST("/networks/:id/rules", s.postNetworkRules())
	er.POST("/networks/:id/devices", s.postDevices())
	er.POST("/networks/:id/devices/:address/authorization", s.postDeviceAuthorization())
	er.POST("/networks/:id/devices/:address/name", s.postDeviceName())
}

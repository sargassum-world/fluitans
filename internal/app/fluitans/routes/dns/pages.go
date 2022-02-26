// Package dns contains the route handlers related to DNS records
package dns

import (
	"github.com/sargassum-eco/fluitans/internal/clients/desec"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/internal/clients/zerotier"
	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/framework"
)

type Service struct {
	r    framework.TemplateRenderer
	dc   *desec.Client
	ztc  *zerotier.Client
	ztcc *ztcontrollers.Client
	sc   *sessions.Client
}

func NewService(
	r framework.TemplateRenderer,
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

func (s *Service) Register(er framework.EchoRouter) {
	er.GET("/dns", s.getServer())
	er.POST("/dns/:subname/:type", s.postRRset())
}

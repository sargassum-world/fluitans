// Package controllers contains the route handlers related to ZeroTier controllers.
package controllers

import (
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/internal/clients/zerotier"
	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/framework"
)

type Service struct {
	r    framework.TemplateRenderer
	ztcc *ztcontrollers.Client
	ztc  *zerotier.Client
	sc   *sessions.Client
}

func NewService(
	r framework.TemplateRenderer,
	ztcc *ztcontrollers.Client, ztc *zerotier.Client, sc *sessions.Client,
) *Service {
	return &Service{
		r:    r,
		ztcc: ztcc,
		ztc:  ztc,
		sc:   sc,
	}
}

func (s *Service) Register(er framework.EchoRouter) {
	er.GET("/controllers", s.getControllers())
	er.GET("/controllers/:name", s.getController())
}

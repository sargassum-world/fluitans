// Package controllers contains the route handlers related to ZeroTier controllers.
package controllers

import (
	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
)

type Handlers struct {
	r godest.TemplateRenderer

	ztcc *ztcontrollers.Client
	ztc  *zerotier.Client
}

func New(
	r godest.TemplateRenderer, ztcc *ztcontrollers.Client, ztc *zerotier.Client,
) *Handlers {
	return &Handlers{
		r:    r,
		ztcc: ztcc,
		ztc:  ztc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter, ss session.Store) {
	hr := auth.NewHTTPRouter(er, ss)
	hr.GET("/controllers", h.HandleControllersGet())
	hr.GET("/controllers/:name", h.HandleControllerGet())
}

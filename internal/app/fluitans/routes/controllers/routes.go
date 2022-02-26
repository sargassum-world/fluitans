// Package controllers contains the route handlers related to ZeroTier controllers.
package controllers

import (
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/internal/clients/zerotier"
	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/godest"
)

type Handlers struct {
	r    godest.TemplateRenderer
	ztcc *ztcontrollers.Client
	ztc  *zerotier.Client
	sc   *sessions.Client
}

func New(
	r godest.TemplateRenderer,
	ztcc *ztcontrollers.Client, ztc *zerotier.Client, sc *sessions.Client,
) *Handlers {
	return &Handlers{
		r:    r,
		ztcc: ztcc,
		ztc:  ztc,
		sc:   sc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	er.GET("/controllers", h.HandleControllersGet())
	er.GET("/controllers/:name", h.HandleControllerGet())
}

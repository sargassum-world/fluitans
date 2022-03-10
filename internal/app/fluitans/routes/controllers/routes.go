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
	r    godest.TemplateRenderer
	ztcc *ztcontrollers.Client
	ztc  *zerotier.Client
	sc   *session.Client
}

func New(
	r godest.TemplateRenderer, ztcc *ztcontrollers.Client, ztc *zerotier.Client, sc *session.Client,
) *Handlers {
	return &Handlers{
		r:    r,
		ztcc: ztcc,
		ztc:  ztc,
		sc:   sc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	ar := auth.NewRouter(er, h.sc)
	ar.GET("/controllers", h.HandleControllersGet())
	ar.GET("/controllers/:name", h.HandleControllerGet())
}

// Package controllers contains the route handlers related to ZeroTier controllers.
package controllers

import (
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/session"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/internal/clients/zerotier"
	"github.com/sargassum-world/fluitans/internal/clients/ztcontrollers"
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

func (h *Handlers) Register(er godest.EchoRouter, ss *session.Store) {
	ar := auth.NewHTTPRouter(er, ss)
	ar.GET("/controllers", h.HandleControllersGet())
	ar.GET("/controllers/:name", h.HandleControllerGet())
}

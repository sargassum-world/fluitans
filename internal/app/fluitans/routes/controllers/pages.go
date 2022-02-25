// Package controllers contains the route handlers related to ZeroTier controllers.
package controllers

import (
	"net/http"

	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/internal/clients/zerotier"
	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

type Service struct {
	ztcc *ztcontrollers.Client
	ztc  *zerotier.Client
	sc   *sessions.Client
}

func NewService(
	ztcc *ztcontrollers.Client, ztc *zerotier.Client, sc *sessions.Client,
) *Service {
	return &Service{
		ztcc: ztcc,
		ztc:  ztc,
		sc:   sc,
	}
}

func (s *Service) Routes() []route.Templated {
	return []route.Templated{
		{
			Path:         "/controllers",
			Method:       http.MethodGet,
			HandlerMaker: s.getControllers,
			Templates:    []string{"controllers/controllers.page.tmpl"},
		},
		{
			Path:         "/controllers/:name",
			Method:       http.MethodGet,
			HandlerMaker: s.getController,
			Templates:    []string{"controllers/controller.page.tmpl"},
		},
	}
}

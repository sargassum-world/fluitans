// Package dns contains the route handlers related to DNS records
package dns

import (
	"net/http"

	"github.com/sargassum-eco/fluitans/internal/clients/desec"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/internal/clients/zerotier"
	"github.com/sargassum-eco/fluitans/internal/clients/ztcontrollers"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

type Service struct {
	dc   *desec.Client
	ztc  *zerotier.Client
	ztcc *ztcontrollers.Client
	sc   *sessions.Client
}

func NewService(
	dc *desec.Client, ztc *zerotier.Client, ztcc *ztcontrollers.Client, sc *sessions.Client,
) *Service {
	return &Service{
		dc:   dc,
		ztc:  ztc,
		ztcc: ztcc,
		sc:   sc,
	}
}

func (s *Service) Routes() []route.Templated {
	return []route.Templated{
		{
			Path:         "/dns",
			Method:       http.MethodGet,
			HandlerMaker: s.getServer,
			Templates:    []string{"dns/server.page.tmpl"},
		},
		{
			Path:         "/dns/:subname/:type",
			Method:       http.MethodPost,
			HandlerMaker: s.postRRset,
		},
	}
}

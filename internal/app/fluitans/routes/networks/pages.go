// Package networks contains the route handlers related to ZeroTier networks.
package networks

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
			Path:         "/networks",
			Method:       http.MethodGet,
			HandlerMaker: s.getNetworks,
			Templates:    []string{"networks/networks.page.tmpl"},
		},
		{
			Path:         "/networks",
			Method:       http.MethodPost,
			HandlerMaker: s.postNetworks,
		},
		{
			Path:         "/networks/:id",
			Method:       http.MethodGet,
			HandlerMaker: s.getNetwork,
			Templates:    []string{"networks/network.page.tmpl"},
		},
		{
			Path:         "/networks/:id",
			Method:       http.MethodPost,
			HandlerMaker: s.postNetwork,
		},
		{
			Path:         "/networks/:id/name",
			Method:       http.MethodPost,
			HandlerMaker: s.postNetworkName,
		},
		{
			Path:         "/networks/:id/rules",
			Method:       http.MethodPost,
			HandlerMaker: s.postNetworkRules,
		},
		{
			Path:         "/networks/:id/devices",
			Method:       http.MethodPost,
			HandlerMaker: s.postDevices,
		},
		{
			Path:         "/networks/:id/devices/:address/authorization",
			Method:       http.MethodPost,
			HandlerMaker: s.postDeviceAuthorization,
		},
		{
			Path:         "/networks/:id/devices/:address/name",
			Method:       http.MethodPost,
			HandlerMaker: s.postDeviceName,
		},
	}
}

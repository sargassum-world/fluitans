// Package networks contains the route handlers related to ZeroTier networks.
package networks

import (
	"net/http"

	"github.com/sargassum-eco/fluitans/internal/route"
)

var Pages = []route.Templated{
	{
		Path:         "/controllers",
		Method:       http.MethodGet,
		HandlerMaker: getControllers,
		Templates:    []string{"networks/controllers.page.tmpl"},
	},
	{
		Path:         "/controllers/:name",
		Method:       http.MethodGet,
		HandlerMaker: getController,
		Templates:    []string{"networks/controller.page.tmpl"},
	},
	{
		Path:         "/networks",
		Method:       http.MethodGet,
		HandlerMaker: getNetworks,
		Templates:    []string{"networks/networks.page.tmpl"},
	},
	{
		Path:         "/networks",
		Method:       http.MethodPost,
		HandlerMaker: postNetworks,
		Templates:    []string{},
	},
	{
		Path:         "/networks/:id",
		Method:       http.MethodGet,
		HandlerMaker: getNetwork,
		Templates:    []string{"networks/network.page.tmpl"},
	},
	{
		Path:         "/networks/:id",
		Method:       http.MethodPost,
		HandlerMaker: postNetwork,
		Templates:    []string{},
	},
	{
		Path:         "/networks/:id/devices",
		Method:       http.MethodPost,
		HandlerMaker: postDevices,
		Templates:    []string{},
	},
	{
		Path:         "/networks/:id/devices/:address",
		Method:       http.MethodPost,
		HandlerMaker: postDevice,
		Templates:    []string{},
	},
}

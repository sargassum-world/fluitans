// Package networks contains the route handlers related to ZeroTier networks.
package networks

import (
	"net/http"

	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

var Pages = []route.Templated{
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
		Path:         "/networks/:id/name",
		Method:       http.MethodPost,
		HandlerMaker: postNetworkName,
		Templates:    []string{},
	},
	{
		Path:         "/networks/:id/rules",
		Method:       http.MethodPost,
		HandlerMaker: postNetworkRules,
		Templates:    []string{},
	},
	{
		Path:         "/networks/:id/devices",
		Method:       http.MethodPost,
		HandlerMaker: postDevices,
		Templates:    []string{},
	},
	{
		Path:         "/networks/:id/devices/:address/authorization",
		Method:       http.MethodPost,
		HandlerMaker: postDeviceAuthorization,
		Templates:    []string{},
	},
	{
		Path:         "/networks/:id/devices/:address/name",
		Method:       http.MethodPost,
		HandlerMaker: postDeviceName,
		Templates:    []string{},
	},
}

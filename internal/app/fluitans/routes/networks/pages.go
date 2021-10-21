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
		HandlerMaker: controllers,
		Templates:    []string{"controllers.page.tmpl"},
	},
	{
		Path:         "/controllers/:name",
		Method:       http.MethodGet,
		HandlerMaker: controller,
		Templates:    []string{"controller.page.tmpl"},
	},
	{
		Path:         "/networks",
		Method:       http.MethodGet,
		HandlerMaker: networks,
		Templates:    []string{"networks.page.tmpl"},
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
		HandlerMaker: network,
		Templates:    []string{"network.page.tmpl"},
	},
	{
		Path:         "/networks/:id",
		Method:       http.MethodPost,
		HandlerMaker: postNetwork,
		Templates:    []string{},
	},
}

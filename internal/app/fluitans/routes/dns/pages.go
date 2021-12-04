// Package dns contains the route handlers related to DNS records
package dns

import (
	"net/http"

	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

var Pages = []route.Templated{
	{
		Path:         "/dns",
		Method:       http.MethodGet,
		HandlerMaker: getServer,
		Templates:    []string{"dns/server.page.tmpl"},
	},
	{
		Path:         "/dns/:subname/:type",
		Method:       http.MethodPost,
		HandlerMaker: postRRset,
		Templates:    []string{},
	},
}

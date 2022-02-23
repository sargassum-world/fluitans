// Package auth contains the route handlers related to authentication and authorization.
package auth

import (
	"net/http"

	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

var Pages = []route.Templated{
	{
		Path:         "/csrf",
		Method:       http.MethodGet,
		HandlerMaker: getCSRF,
		Templates:    []string{},
	},
	{
		Path:         "/login",
		Method:       http.MethodGet,
		HandlerMaker: getLogin,
		Templates:    []string{"auth/login.page.tmpl"},
	},
	{
		Path:         "/sessions",
		Method:       http.MethodPost,
		HandlerMaker: postSessions,
		Templates:    []string{},
	},
}

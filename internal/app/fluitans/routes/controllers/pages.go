// Package controllers contains the route handlers related to ZeroTier controllers.
package controllers

import (
	"net/http"

	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

var Pages = []route.Templated{
	{
		Path:         "/controllers",
		Method:       http.MethodGet,
		HandlerMaker: getControllers,
		Templates:    []string{"controllers/controllers.page.tmpl"},
	},
	{
		Path:         "/controllers/:name",
		Method:       http.MethodGet,
		HandlerMaker: getController,
		Templates:    []string{"controllers/controller.page.tmpl"},
	},
}

// Package assets contains the route handlers for assets which are static for the server
package assets

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/pkg/framework/embeds"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

type TemplatedService struct{}

func NewTemplatedService() *TemplatedService {
	return &TemplatedService{}
}

func (s *TemplatedService) Routes() []route.Templated {
	return []route.Templated{
		{
			Path:         "/app/app.webmanifest",
			Method:       http.MethodGet,
			HandlerMaker: getWebmanifest,
			Templates:    []string{"app/app.webmanifest.tmpl"},
		},
	}
}

func RegisterStatic(er route.EchoRouter, em embeds.Embeds) {
	const (
		day  = 24 * time.Hour
		week = 7 * day
		year = 365 * day
	)

	er.GET("/favicon.ico", echo.WrapHandler(route.HandleFS("/", em.StaticFS, week)))
	er.GET("/fonts/*", echo.WrapHandler(route.HandleFS("/fonts/", em.FontsFS, year)))
	er.GET("/static/*", echo.WrapHandler(route.HandleFSFileRevved("/static/", em.StaticHFS)))
	er.GET("/app/*", echo.WrapHandler(route.HandleFSFileRevved("/app/", em.AppHFS)))
}

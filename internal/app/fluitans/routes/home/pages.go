// Package home contains the route handlers related to the app's home screen.
package home

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

type Service struct {
	sc *sessions.Client
}

func NewService(sc *sessions.Client) *Service {
	return &Service{
		sc: sc,
	}
}

func (s *Service) Routes() []route.Templated {
	return []route.Templated{
		{
			Path:         "/",
			Method:       http.MethodGet,
			HandlerMaker: s.getHome,
			Templates:    []string{"home/home.page.tmpl"},
		},
	}
}

func (s *Service) getHome(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "home/home.page.tmpl"
	te.Require(t)
	return func(c echo.Context) error {
		// Check authentication & authorization
		a, _, err := auth.GetWithSession(c, s.sc)
		if err != nil {
			return err
		}

		// Produce output
		return route.Render(c, t, struct{}{}, a, te, g)
	}, nil
}

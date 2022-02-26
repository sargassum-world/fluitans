// Package home contains the route handlers related to the app's home screen.
package home

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/pkg/godest"
)

type Service struct {
	r  godest.TemplateRenderer
	sc *sessions.Client
}

func NewService(r godest.TemplateRenderer, sc *sessions.Client) *Service {
	return &Service{
		r:  r,
		sc: sc,
	}
}

func (s *Service) Register(er godest.EchoRouter) {
	er.GET("/", s.getHome())
}

func (s *Service) getHome() echo.HandlerFunc {
	t := "home/home.page.tmpl"
	s.r.MustHave(t)
	return func(c echo.Context) error {
		// Check authentication & authorization
		a, _, err := auth.GetWithSession(c, s.sc)
		if err != nil {
			return err
		}

		// Produce output
		return s.r.CacheablePage(c.Response(), c.Request(), t, struct{}{}, a)
	}
}

package controllers

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
)

func (s *Service) getControllers() echo.HandlerFunc {
	t := "controllers/controllers.page.tmpl"
	s.r.MustHave(t)
	return func(c echo.Context) error {
		// Check authentication & authorization
		a, _, err := auth.GetWithSession(c, s.sc)
		if err != nil {
			return err
		}

		// Run queries
		controllers, err := s.ztcc.GetControllers()
		if err != nil {
			return err
		}

		// Produce output
		return s.r.CacheablePage(c.Response(), c.Request(), t, controllers, a)
	}
}

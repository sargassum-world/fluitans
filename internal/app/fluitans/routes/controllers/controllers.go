package controllers

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

func (s *Service) getControllers(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	t := "controllers/controllers.page.tmpl"
	te.Require(t)
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
		return route.Render(c, t, controllers, a, te, g)
	}, nil
}

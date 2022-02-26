package controllers

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
)

func (h *Handlers) HandleControllersGet() echo.HandlerFunc {
	t := "controllers/controllers.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context) error {
		// Check authentication & authorization
		a, _, err := auth.GetWithSession(c, h.sc)
		if err != nil {
			return err
		}

		// Run queries
		controllers, err := h.ztcc.GetControllers()
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, controllers, a)
	}
}

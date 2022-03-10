package controllers

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
)

func (h *Handlers) HandleControllersGet() auth.Handler {
	t := "controllers/controllers.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Run queries
		controllers, err := h.ztcc.GetControllers()
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, controllers, a)
	}
}

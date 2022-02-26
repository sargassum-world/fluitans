// Package home contains the route handlers related to the app's home screen.
package home

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/pkg/godest"
)

type Handlers struct {
	r  godest.TemplateRenderer
	sc *sessions.Client
}

func New(r godest.TemplateRenderer, sc *sessions.Client) *Handlers {
	return &Handlers{
		r:  r,
		sc: sc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	er.GET("/", h.HandleHomeGet())
}

func (h *Handlers) HandleHomeGet() echo.HandlerFunc {
	t := "home/home.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context) error {
		// Check authentication & authorization
		a, _, err := auth.GetWithSession(c, h.sc)
		if err != nil {
			return err
		}

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, struct{}{}, a)
	}
}

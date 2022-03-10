// Package home contains the route handlers related to the app's home screen.
package home

import (
	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/pkg/godest"
	"github.com/sargassum-world/fluitans/pkg/godest/session"
)

type Handlers struct {
	r  godest.TemplateRenderer
	sc *session.Client
}

func New(r godest.TemplateRenderer, sc *session.Client) *Handlers {
	return &Handlers{
		r:  r,
		sc: sc,
	}
}

func (h *Handlers) Register(er godest.EchoRouter) {
	ar := auth.NewRouter(er, h.sc)
	ar.GET("/", h.HandleHomeGet())
}

func (h *Handlers) HandleHomeGet() auth.Handler {
	t := "home/home.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth) error {
		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, struct{}{}, a)
	}
}

// Package routes contains the route handlers for the Fluitans server.
package routes

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/caching"
	"github.com/sargassum-eco/fluitans/internal/fingerprint"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
)

var Pages = []route.Templated{
	{
		Path:         "/",
		Method:       http.MethodGet,
		HandlerMaker: home,
		Templates:    []string{"home.page.tmpl"},
	},
	{
		Path:         "/login",
		Method:       http.MethodGet,
		HandlerMaker: login,
		Templates:    []string{"login.page.tmpl"},
	},
	{
		Path:         "/hello/:name",
		Method:       http.MethodGet,
		HandlerMaker: helloName,
		Templates:    []string{"hello.page.tmpl"},
	},
}

func home(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "home.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	return func(c echo.Context) error {
		// Handle Etag
		if noContent, err := caching.ProcessEtag(c, tte); noContent {
			return err
		}

		// Render template
		return c.Render(http.StatusOK, t, struct {
			Meta   template.Meta
			Embeds template.Embeds
		}{
			Meta: template.Meta{
				Description: "An application for managing Sargassum networks, domain names, and organizations.",
				Path:        c.Request().URL.Path,
			},
			Embeds: g.Embeds,
		})
	}, nil
}

func login(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "login.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	return func(c echo.Context) error {
		// Handle Etag
		if noContent, err := caching.ProcessEtag(c, tte); noContent {
			return err
		}

		// Render template
		return c.Render(http.StatusOK, t, struct {
			Meta   template.Meta
			Embeds template.Embeds
		}{
			Meta: template.Meta{
				Description: "Log in to Fluitans.",
				Path:        c.Request().URL.Path,
			},
			Embeds: g.Embeds,
		})
	}, nil
}

func helloName(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "hello.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	return func(c echo.Context) error {
		// Parse params
		name := c.Param("name")

		// Handle Etag
		if noContent, err := caching.ProcessEtag(
			c, tte, fingerprint.Compute([]byte(name)),
		); noContent {
			return err
		}

		// Render template
		return c.Render(http.StatusOK, t, struct {
			Meta   template.Meta
			Embeds template.Embeds
			Name   string
		}{
			Meta: template.Meta{
				Description: fmt.Sprintf("A greeting for %s.", name),
				Path:        c.Request().URL.Path,
			},
			Embeds: g.Embeds,
			Name:   name,
		})
	}, nil
}

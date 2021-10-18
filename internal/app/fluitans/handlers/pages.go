package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/templates"
)

func home(c echo.Context) error {
	return c.Render(http.StatusOK, "home.page.tmpl", struct {
		Meta        templates.Meta
		EmbedAssets templates.EmbedAssets
	}{
		Meta: templates.Meta{
			Description: "An application for managing Sargassum networks, domain names, and organizations.",
			Path:        c.Request().URL.Path,
		},
		EmbedAssets: embedAssets,
	})
}

func login(c echo.Context) error {
	return c.Render(http.StatusOK, "login.page.tmpl", struct {
		Meta        templates.Meta
		EmbedAssets templates.EmbedAssets
	}{
		Meta: templates.Meta{
			Description: "Log in to Fluitans.",
			Path:        c.Request().URL.Path,
		},
		EmbedAssets: embedAssets,
	})
}

func helloName(c echo.Context) error {
	name := c.Param("name")
	return c.Render(http.StatusOK, "hello.page.tmpl", struct {
		Meta        templates.Meta
		EmbedAssets templates.EmbedAssets
		Name        string
	}{
		Meta: templates.Meta{
			Description: fmt.Sprintf("A greeting for %s.", name),
			Path:        c.Request().URL.Path,
		},
		EmbedAssets: embedAssets,
		Name:        name,
	})
}

// Package templates contains the templates to be rendered by the Fluitans server.
package templates

import (
	"fmt"
	"html/template"
	"io"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/web"
)

type TemplateRenderer struct {
	templates *template.Template
}

func New() *TemplateRenderer {
	return &TemplateRenderer{
		templates: template.Must(
			template.
				New("").
				Funcs(sprig.FuncMap()).
				Funcs(template.FuncMap{
					"appHashed": func(file string) string {
						return fmt.Sprintf("/app/%s", web.AppHFS.HashName(file))
					},
					"staticHashed": func(file string) string {
						return fmt.Sprintf("/static/%s", web.StaticHFS.HashName(file))
					},
				}).
				ParseFS(web.TemplatesFS, "*.tmpl", "*/*.tmpl"),
		),
	}
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl := template.Must(t.templates.Clone())
	tmpl = template.Must(tmpl.ParseFS(web.TemplatesFS, name))
	return tmpl.ExecuteTemplate(w, name, data)
}

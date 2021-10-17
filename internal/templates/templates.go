// Package templates contains the templates to be rendered by the Fluitans server.
package templates

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/Masterminds/sprig/v3"

	"github.com/sargassum-eco/fluitans/web"
)

type TemplateRenderer struct {
	templates   *template.Template
}

func New() *TemplateRenderer {
	return &TemplateRenderer{
		templates:   template.Must(template.New("").Funcs(sprig.FuncMap()).ParseFS(web.TemplatesFS, "*.tmpl", "*/*.tmpl")),
	}
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl := template.Must(t.templates.Clone())
	tmpl = template.Must(tmpl.ParseFS(web.TemplatesFS, name))
	return tmpl.ExecuteTemplate(w, name, data)
}

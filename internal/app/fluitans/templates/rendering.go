// Package templates contains the templates to be rendered by the Fluitans server.
package templates

import (
	"html/template"
	"io"
	"io/fs"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"
)

type TemplateRenderer struct {
	templates   *template.Template
	templatesFS fs.FS
}

func New(appNamer, staticNamer HashNamer, templates fs.FS) *TemplateRenderer {
	return &TemplateRenderer{
		templates: template.Must(
			template.
				New("App").
				Funcs(sprig.FuncMap()).
				Funcs(template.FuncMap{
					"appHashed":    getHashedName("app", appNamer),
					"staticHashed": getHashedName("static", staticNamer),
				}).
				ParseFS(templates, "*.tmpl", "*/*.tmpl"),
		),
		templatesFS: templates,
	}
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl := template.Must(t.templates.Clone())
	tmpl = template.Must(tmpl.ParseFS(t.templatesFS, name))
	return tmpl.ExecuteTemplate(w, name, data)
}

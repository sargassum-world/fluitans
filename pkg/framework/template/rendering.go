package template

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

func (t *TemplateRenderer) Render(
	w io.Writer, name string, data interface{}, c echo.Context,
) error {
	tmpl := template.Must(t.templates.Clone())
	tmpl = template.Must(ParseFS(tmpl, t.templatesFS, name))
	return tmpl.ExecuteTemplate(w, name, data)
}

func NewTemplateRenderer(fsys fs.FS, functions ...template.FuncMap) *TemplateRenderer {
	tmpl := template.New("App").Funcs(sprig.FuncMap())
	for _, f := range functions {
		tmpl = tmpl.Funcs(f)
	}
	return &TemplateRenderer{
		templates:   template.Must(ParseFS(tmpl, fsys, "**/*.tmpl")),
		templatesFS: fsys,
	}
}

// Package templates contains the templates to be rendered by the Fluitans server.
package templates

import (
	"html/template"
	"io"
	"io/fs"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"

	tp "github.com/sargassum-eco/fluitans/internal/template"
)

type TemplateRenderer struct {
	templates   *template.Template
	templatesFS fs.FS
}

func New(appNamer, staticNamer HashNamer, fsys fs.FS) *TemplateRenderer {
	tmpl :=
		template.
			New("App").
			Funcs(sprig.FuncMap()).
			Funcs(template.FuncMap{
				"appHashed":              getHashedName("app", appNamer),
				"staticHashed":           getHashedName("static", staticNamer),
				"describeError":          describeError,
				"identifyNetwork":        identifyNetwork,
				"getNetworkHostAddress":  getNetworkHostAddress,
				"getNetworkNumber":       getNetworkNumber,
				"durationToSec":          durationToSec,
				"derefBool":              derefBool,
				"derefInt":               derefInt,
				"derefFloat32":           derefFloat32,
				"describeDNSRecordType":  describeDNSRecordType,
				"exemplifyDNSRecordType": exemplifyDNSRecordType,
			})
	return &TemplateRenderer{
		templates: template.Must(
			tp.ParseFS(tmpl, fsys, "*.tmpl", "*/*.tmpl"),
		),
		templatesFS: fsys,
	}
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl := template.Must(t.templates.Clone())
	tmpl = template.Must(tp.ParseFS(tmpl, t.templatesFS, name))
	return tmpl.ExecuteTemplate(w, name, data)
}

// Package templates contains the templates to be rendered by the Fluitans server.
package templates

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"

	"github.com/Masterminds/sprig/v3"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/route"
	tp "github.com/sargassum-eco/fluitans/internal/template"
)

type TemplateRenderer struct {
	templates   *template.Template
	templatesFS fs.FS
}

func (t *TemplateRenderer) Render(
	w io.Writer, name string, data interface{}, c echo.Context,
) error {
	tmpl := template.Must(t.templates.Clone())
	tmpl = template.Must(tp.ParseFS(tmpl, t.templatesFS, name))
	return tmpl.ExecuteTemplate(w, name, data)
}

func FuncMap(appNamer, staticNamer HashNamer) template.FuncMap {
	return template.FuncMap{
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
		"derefString":            derefString,
		"describeDNSRecordType":  describeDNSRecordType,
		"exemplifyDNSRecordType": exemplifyDNSRecordType,
	}
}

func NewRenderer(functions template.FuncMap, fsys fs.FS) *TemplateRenderer {
	tmpl := template.New("App").Funcs(sprig.FuncMap()).Funcs(functions)
	return &TemplateRenderer{
		templates:   template.Must(tp.ParseFS(tmpl, fsys, "*.tmpl", "*/*.tmpl")),
		templatesFS: fsys,
	}
}

func GetTemplate(te route.TemplateEtagSegments, name, route string) ([]string, error) {
	tte, ok := te[name]
	if !ok {
		return nil, errors.Wrap(
			te.NewNotFoundError(name), fmt.Sprintf("couldn't find template for %s", route),
		)
	}
	return tte, nil
}

type RenderData struct {
	Meta   tp.Meta
	Embeds tp.Embeds
	Data   interface{}
}

func MakeRenderData(
	c echo.Context, g route.TemplateGlobals, data interface{},
) RenderData {
	return RenderData{
		Meta: tp.Meta{
			Path: c.Request().URL.Path,
		},
		Embeds: g.Embeds,
		Data:   data,
	}
}

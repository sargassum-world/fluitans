// Package framework provides a reusable framework for Fluitans-style web apps
package framework

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework/httpcache"
	tp "github.com/sargassum-eco/fluitans/pkg/framework/template"
)

// Template rendering data

type Meta struct {
	Path       string
	RequestURI string
}

type RenderData struct {
	Meta    Meta
	Inlines interface{}
	Data    interface{}
	Auth    interface{}
}

type TemplateRenderer struct {
	inlines      interface{}
	fingerprints Fingerprints
	templates    tp.Templates
}

func NewTemplateRenderer(
	e Embeds, inlines interface{}, funcs ...template.FuncMap,
) (r TemplateRenderer, err error) {
	r.inlines = inlines
	r.templates, err = e.NewTemplates(funcs...)
	if err != nil {
		err = errors.Wrap(err, "couldn't make templated pages renderer")
		return
	}
	if r.fingerprints.App, err = e.ComputeAppFingerprint(); err != nil {
		err = errors.Wrap(err, "couldn't compute fingerprint for app")
		return
	}
	if r.fingerprints.Page, err = e.ComputePageFingerprints(); err != nil {
		err = errors.Wrap(err, "couldn't compute fingerprint for page/module templates")
		return
	}
	return
}

func (tr TemplateRenderer) GetEchoRenderer() EchoRenderer {
	return NewEchoRenderer(tr.templates)
}

func (tr TemplateRenderer) MustHave(templateNames ...string) {
	tr.fingerprints.MustHave(templateNames...)
}

func (tr TemplateRenderer) NewRenderData(
	r *http.Request, data interface{}, auth interface{},
) RenderData {
	return RenderData{
		Meta: Meta{
			Path:       r.URL.Path,
			RequestURI: r.URL.RequestURI(),
		},
		Inlines: tr.inlines,
		Data:    data,
		Auth:    auth,
	}
}

func (tr TemplateRenderer) SetUncacheable(resh http.Header) {
	httpcache.SetNoEtag(resh)
}

func (tr TemplateRenderer) Page(
	w http.ResponseWriter, r *http.Request,
	status int, templateName string, templateData interface{}, authData interface{},
) error {
	// This is basically a reimplementation of the echo.Context.Render method, but without requiring
	// an echo.Context to be provided
	buf := new(bytes.Buffer)
	if rerr := tr.templates.Execute(
		buf, templateName, tr.NewRenderData(r, templateData, authData),
	); rerr != nil {
		return rerr
	}

	// Write render result
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, werr := w.Write(buf.Bytes())
	return werr
}

func (tr TemplateRenderer) CacheablePage(
	w http.ResponseWriter, r *http.Request,
	templateName string, templateData interface{}, authData interface{},
) error {
	type EtagInputs struct {
		Data interface{}
		Auth interface{}
	}
	if noContent, err := tr.fingerprints.SetAndCheckEtag(w, r, templateName, EtagInputs{
		Data: templateData,
		Auth: authData,
	}); noContent || (err != nil) {
		return err
	}
	return tr.Page(w, r, http.StatusOK, templateName, templateData, authData)
}

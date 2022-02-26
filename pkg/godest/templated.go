// Package godest provides a mildly opinionated framework for more easily writing web apps with
// modest Javascript approaches such as Hotwire-based apps. It provides support for using templates
// in a clear and structured way. It also makes it easy to embed all templates, static assets
// (e.g. images) and app-related assets (e.g. JS bundles) into the compiled server binary, and it
// takes care of the details of browser caching for assets and templated pages.
// Finally, it provides some standalone utilities for caching data on the server, getting the values
// of environment variables, and using cookies for form-based authentication.
package godest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/godest/fingerprint"
	"github.com/sargassum-eco/fluitans/pkg/godest/httpcache"
	tp "github.com/sargassum-eco/fluitans/pkg/godest/template"
)

// Templated page fingerprinting

type fingerprints struct {
	App  string
	Page map[string]string
}

func (f fingerprints) GetEtagSegments(templateName string) ([]string, error) {
	if templateName == "" {
		return []string{f.App}, nil
	}

	pageFingerprint, ok := f.Page[templateName]
	if !ok {
		return []string{f.App}, errors.Errorf(
			"couldn't find page fingerprint for template %s", templateName,
		)
	}

	return []string{f.App, pageFingerprint}, nil
}

func (f fingerprints) MustHave(templateNames ...string) {
	for _, name := range templateNames {
		if _, err := f.GetEtagSegments(name); err != nil {
			panic(errors.Wrap(err, fmt.Sprintf("couldn't find template etag segments for %s", name)))
		}
	}
}

func (f fingerprints) SetAndCheckEtag(
	w http.ResponseWriter, r *http.Request, templateName string, data interface{},
) (noContent bool, err error) {
	// Look up data-independent etag segments
	templateEtagSegments, err := f.GetEtagSegments(templateName)
	if err != nil {
		return
	}

	// Encode data
	var buf bytes.Buffer
	// github.com/vmihailenco/msgpack has better performance, but we use the JSON encoder because
	// the msgpack encoder can only sort the map keys of map[string]string and map[string]interface{}
	// maps, and it's too much trouble to convert our maps into map[string]interface{}. If we can
	// work around this limitation, we should use msgpack though.
	if err = json.NewEncoder(&buf).Encode(data); err != nil {
		return
	}
	encoded := buf.Bytes()

	noContent = httpcache.SetAndCheckEtag(
		w, r, append(templateEtagSegments, fingerprint.Compute(encoded))...,
	)
	return
}

// Template rendering

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
	fingerprints fingerprints
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

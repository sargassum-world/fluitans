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
	"html/template"
	"io/fs"
	"net/http"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/pkg/errors"
)

// Template File Filters

const (
	templateFileExt        = ".tmpl"
	partialTemplateFileExt = ".partial" + templateFileExt
	layoutTemplateFileExt  = ".layout" + templateFileExt
	templateSharedModule   = "shared"
)

func filterTemplate(path string) bool {
	return strings.HasSuffix(path, templateFileExt)
}

func filterSharedTemplate(path string) bool {
	return filterTemplate(path) && strings.HasPrefix(path, templateSharedModule+"/")
}

func filterPartialTemplate(path string) bool {
	return strings.HasSuffix(path, partialTemplateFileExt)
}

func filterLayoutTemplate(path string) bool {
	return strings.HasSuffix(path, layoutTemplateFileExt)
}

func filterNonpageTemplate(path string) bool {
	return filterPartialTemplate(path) || filterLayoutTemplate(path)
}

func filterPageTemplate(path string) bool {
	// Usually a page ends with ".page.tmpl", but it may end with other things,
	// e.g. ".webmanifest.tmpl", ".json.tmpl", etc.
	return filterTemplate(path) && !filterNonpageTemplate(path)
}

func filterTemplateModule(path string) bool {
	return path != templateSharedModule
}

// Built Asset File Filters

func filterCSSAsset(path string) bool {
	return strings.HasSuffix(path, ".css")
}

func filterJSAsset(path string) bool {
	return strings.HasSuffix(path, ".js")
}

func filterAsset(path string) bool {
	return filterCSSAsset(path) || filterJSAsset(path)
}

// Template File Parsing

// parseFiles is a direct copy of the parseFiles helper function in text/template.
func parseFiles(
	t *template.Template, readFile func(string) (string, []byte, error), filenames ...string,
) (*template.Template, error) {
	if len(filenames) == 0 {
		return nil, errors.Errorf("template: no files named in call to ParseFiles")
	}

	for _, filename := range filenames {
		name, b, err := readFile(filename)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't read file %s", filename)
		}
		s := string(b)
		var tmpl *template.Template
		if t == nil {
			t = template.New(name)
		}
		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't parse template %s", filename)
		}
	}
	return t, nil
}

// readFileFS is an adaptation of the readFileFS helper function in text/template, except it uses
// the fully-qualified path name of the template file in the filesystem rather than just the
// template file's basename.
func readFileFS(fsys fs.FS) func(string) (string, []byte, error) {
	return func(file string) (name string, b []byte, err error) {
		name = file
		b, err = fs.ReadFile(fsys, file)
		return
	}
}

// parseFS is a direct copy of the parseFS helper function in text/template, except it uses the
// fully-qualified path name of the template file in the filesystem rather than just the template
// file's basename, and it allows double-star globs (e.g. "**/*.txt").
func parseFS(
	t *template.Template, fsys fs.FS, patterns ...string,
) (*template.Template, error) {
	var filenames []string
	for _, pattern := range patterns {
		list, err := doublestar.Glob(fsys, pattern)
		if err != nil {
			return nil, err
		}
		if len(list) == 0 {
			return nil, errors.Errorf("template: pattern matches no files: %#q", pattern)
		}
		filenames = append(filenames, list...)
	}
	return parseFiles(t, readFileFS(fsys), filenames...)
}

// Templated Page Fingerprinting

type fingerprints struct {
	App  string
	Page map[string]string
}

func (f fingerprints) getEtagSegments(templateName string) ([]string, error) {
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
		if _, err := f.getEtagSegments(name); err != nil {
			panic(errors.Wrapf(err, "couldn't find template etag segments for %s", name))
		}
	}
}

func (f fingerprints) SetAndCheckEtag(
	w http.ResponseWriter, r *http.Request, templateName string, data interface{},
) (noContent bool, err error) {
	// Look up data-independent etag segments
	templateEtagSegments, err := f.getEtagSegments(templateName)
	if err != nil {
		return false, err
	}

	// Encode data
	var buf bytes.Buffer
	// github.com/vmihailenco/msgpack has better performance, but we use the JSON encoder because
	// the msgpack encoder can only sort the map keys of map[string]string and map[string]interface{}
	// maps, and it's too much trouble to convert our maps into map[string]interface{}. If we can
	// work around this limitation, we should use msgpack though.
	if err = json.NewEncoder(&buf).Encode(data); err != nil {
		return false, err
	}
	encoded := buf.Bytes()

	noContent = setAndCheckEtag(
		w, r, append(templateEtagSegments, computeFingerprint(encoded))...,
	)
	if noContent {
		w.WriteHeader(http.StatusNotModified)
	}
	return noContent, nil
}

// Template Rendering

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
	pageTemplates map[string]*template.Template
	inlines       interface{}
	fingerprints  fingerprints
}

func NewTemplateRenderer(
	e Embeds, inlines interface{}, funcs ...template.FuncMap,
) (tr TemplateRenderer, err error) {
	tmpl := template.New("App")
	for _, f := range funcs {
		tmpl = tmpl.Funcs(f)
	}
	allTemplates, err := parseFS(tmpl, e.TemplatesFS, "**/*"+templateFileExt)
	if err != nil {
		return TemplateRenderer{}, errors.Wrap(err, "couldn't load templates from filesystem")
	}

	tr.pageTemplates = make(map[string]*template.Template)
	pageFiles, err := listFiles(e.TemplatesFS, filterPageTemplate)
	if err != nil {
		return TemplateRenderer{}, errors.Wrap(err, "couldn't list template pages")
	}
	for _, name := range pageFiles {
		all, cerr := allTemplates.Clone()
		if cerr != nil {
			return TemplateRenderer{}, errors.Wrap(err, "couldn't clone root template set")
		}
		tr.pageTemplates[name], err = parseFS(all, e.TemplatesFS, name)
		if err != nil {
			return TemplateRenderer{}, errors.Wrapf(err, "couldn't make template set for %s", name)
		}
	}

	tr.inlines = inlines
	if tr.fingerprints.App, err = e.computeAppFingerprint(); err != nil {
		return TemplateRenderer{}, errors.Wrap(err, "couldn't compute fingerprint for app")
	}
	if tr.fingerprints.Page, err = e.computePageFingerprints(); err != nil {
		return TemplateRenderer{}, errors.Wrap(
			err, "couldn't compute fingerprint for page/module templates",
		)
	}

	return tr, nil
}

func (tr TemplateRenderer) MustHave(templateNames ...string) {
	tr.fingerprints.MustHave(templateNames...)
}

func (tr TemplateRenderer) newRenderData(
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

func (tr TemplateRenderer) Page(
	w http.ResponseWriter, r *http.Request,
	status int, templateName string, templateData interface{}, authData interface{},
) error {
	// This is basically a reimplementation of the echo.Context.Render method, but without requiring
	// an echo.Context to be provided
	buf := new(bytes.Buffer)
	tmpl, ok := tr.pageTemplates[templateName]
	if !ok {
		return errors.Errorf("teplate %s not found", templateName)
	}
	if err := tmpl.ExecuteTemplate(
		buf, templateName, tr.newRenderData(r, templateData, authData),
	); err != nil {
		return errors.Wrapf(err, "couldn't execute template %s", templateName)
	}

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

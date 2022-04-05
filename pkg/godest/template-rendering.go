// Package godest provides a mildly opinionated framework for more easily writing web apps with
// modest Javascript approaches, especially with the Hotwire libraries. It provides support for
// using templates in a clear and structured way. It also makes it easy to embed all templates,
// static assets (e.g. images) and app-related assets (e.g. JS bundles) into the compiled server
// binary, and it takes care of the details of browser caching for assets and templated pages.
// Finally, it provides some standalone utilities for caching data on the server, getting the values
// of environment variables, and using cookies for form-based authentication.
package godest

import (
	"bytes"
	"html/template"
	"io"
	"io/fs"
	"net/http"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"
)

type RenderData struct {
	Meta struct {
		Path       string
		RequestURI string
	}
	Inlines interface{}
	Data    interface{}
	Auth    interface{}
}

type TemplateRenderer struct {
	allTemplates         *template.Template
	partialTemplates     map[string]*template.Template
	pageTemplates        map[string]*template.Template
	turboStreamsTemplate *template.Template
	inlines              interface{}
	fingerprints         fingerprints
}

func instantiateTemplates(
	templatesFS fs.FS, allTemplates *template.Template, templateFilter func(string) bool,
) (templates map[string]*template.Template, err error) {
	templates = make(map[string]*template.Template)
	templateFiles, err := listFiles(templatesFS, templateFilter)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't list templates")
	}
	for _, name := range templateFiles {
		all, err := allTemplates.Clone()
		if err != nil {
			return nil, errors.Wrap(err, "couldn't clone root template set")
		}
		templates[name], err = parseFS(all, templatesFS, name)
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't make template set for %s", name)
		}
	}
	return templates, nil
}

func instantiateTurboStreamsTemplate(
	allTemplates *template.Template, turboStreamsTemplate string,
) (t *template.Template, err error) {
	all, err := allTemplates.Clone()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't clone root template set")
	}
	t, err = all.Parse(turboStreamsTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't parse Turbo Streams template")
	}
	return t, nil
}

func NewTemplateRenderer(
	e Embeds, inlines interface{}, funcs ...template.FuncMap,
) (tr TemplateRenderer, err error) {
	tmpl := template.New("App")
	for _, f := range funcs {
		tmpl = tmpl.Funcs(f)
	}
	tr.allTemplates, err = parseFS(tmpl, e.TemplatesFS, "**/*"+templateFileExt)
	if err != nil {
		return TemplateRenderer{}, errors.Wrap(err, "couldn't load templates from filesystem")
	}

	tr.pageTemplates, err = instantiateTemplates(
		e.TemplatesFS, tr.allTemplates, filterPageTemplate,
	)
	if err != nil {
		return TemplateRenderer{}, errors.Wrap(err, "couldn't instantiate page templates")
	}
	tr.partialTemplates, err = instantiateTemplates(
		e.TemplatesFS, tr.allTemplates, filterPartialTemplate,
	)
	if err != nil {
		return TemplateRenderer{}, errors.Wrap(err, "couldn't instantiate partial templates")
	}
	tr.turboStreamsTemplate, err = instantiateTurboStreamsTemplate(
		tr.allTemplates, turbostreams.Template,
	)
	if err != nil {
		return TemplateRenderer{}, errors.Wrap(err, "couldn't instantiate Turbo Streams template")
	}

	tr.inlines = inlines
	if tr.fingerprints.app, err = e.computeAppFingerprint(); err != nil {
		return TemplateRenderer{}, errors.Wrap(err, "couldn't compute fingerprint for app")
	}
	if tr.fingerprints.page, err = e.computePageFingerprints(); err != nil {
		return TemplateRenderer{}, errors.Wrap(
			err, "couldn't compute fingerprint for page/module templates",
		)
	}

	return tr, nil
}

func (tr TemplateRenderer) MustHave(templateNames ...string) {
	for _, templateName := range templateNames {
		if tr.allTemplates.Lookup(templateName) == nil {
			panic(errors.Errorf("couldn't find required template %s", templateName))
		}
		if filterPageTemplate(templateName) {
			tr.fingerprints.mustHaveForPage(templateName)
		}
	}
}

func (tr TemplateRenderer) newRenderData(
	r *http.Request, data interface{}, auth interface{},
) RenderData {
	return RenderData{
		Meta: struct {
			Path       string
			RequestURI string
		}{
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
	headerOptions ...HeaderOption,
) error {
	// This is basically a reimplementation of the echo.Context.Render method, but without requiring
	// an echo.Context to be provided
	buf := new(bytes.Buffer)
	tmpl, ok := tr.pageTemplates[templateName]
	if !ok {
		return errors.Errorf("page template %s not found", templateName)
	}
	if err := tmpl.ExecuteTemplate(
		buf, templateName, tr.newRenderData(r, templateData, authData),
	); err != nil {
		return errors.Wrapf(err, "couldn't execute page template %s", templateName)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	for _, headerOption := range headerOptions {
		headerOption(w.Header())
	}
	w.WriteHeader(status)
	_, werr := w.Write(buf.Bytes())
	return werr
}

func (tr TemplateRenderer) CacheablePage(
	w http.ResponseWriter, r *http.Request,
	templateName string, templateData interface{}, authData interface{},
	headerOptions ...HeaderOption,
) error {
	type EtagInputs struct {
		Data interface{}
		Auth interface{}
	}
	if noContent, err := tr.fingerprints.setAndCheckEtag(w, r, templateName, EtagInputs{
		Data: templateData,
		Auth: authData,
	}); noContent || (err != nil) {
		return err
	}
	return tr.Page(w, r, http.StatusOK, templateName, templateData, authData, headerOptions...)
}

func (tr TemplateRenderer) WritePartial(
	w io.Writer, partialName string, partialData interface{},
) error {
	tmpl, ok := tr.partialTemplates[partialName]
	if !ok {
		return errors.Errorf("partial template %s not found", partialName)
	}
	if err := tmpl.ExecuteTemplate(w, partialName, partialData); err != nil {
		return errors.Wrapf(err, "couldn't execute partial template %s", partialName)
	}
	return nil
}

type renderedStreamMessage struct {
	Action   turbostreams.Action
	Targets  string
	Target   string
	Rendered template.HTML
}

func (tr TemplateRenderer) WriteTurboStream(w io.Writer, messages ...turbostreams.Message) error {
	rendered := make([]renderedStreamMessage, len(messages))
	for i, stream := range messages {
		buf := new(bytes.Buffer)
		if err := tr.WritePartial(buf, stream.Template, stream.Data); err != nil {
			return errors.Wrapf(err, "couldn't execute stream template %s", stream.Template)
		}
		rendered[i] = renderedStreamMessage{
			Action:  stream.Action,
			Targets: stream.Targets,
			Target:  stream.Target,
			//nolint:gosec // This is generated from trusted templates, so we know it's well-formed
			Rendered: template.HTML(buf.String()),
		}
	}

	buf := new(bytes.Buffer)
	if err := tr.turboStreamsTemplate.Execute(buf, rendered); err != nil {
		return errors.Wrap(err, "couldn't execute Turbo Streams template")
	}

	_, werr := w.Write(buf.Bytes())
	return werr
}

func (tr TemplateRenderer) TurboStream(
	w http.ResponseWriter, messages ...turbostreams.Message,
) error {
	w.Header().Set("Content-Type", turbostreams.ContentType+"; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	return tr.WriteTurboStream(w, messages...)
}

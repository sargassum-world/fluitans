package route

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework/template"
)

// Template rendering

type Meta struct {
	Path       string
	RequestURI string
}

type RenderData struct {
	Meta    Meta
	Inlines Inlines
	Data    interface{}
	Auth    interface{}
}

func NewRenderData(
	r *http.Request, g TemplateGlobals, data interface{}, auth interface{},
) RenderData {
	return RenderData{
		Meta: Meta{
			Path:       r.URL.Path,
			RequestURI: r.URL.RequestURI(),
		},
		Inlines: g.Inlines,
		Data:    data,
		Auth:    auth,
	}
}

func WriteTemplatedResponse(
	w http.ResponseWriter, r *http.Request, renderer *template.TemplateRenderer,
	templateName string, status int, templateData interface{}, authData interface{},
	g TemplateGlobals,
) error {
	// This is basically a reimplementation of the echo.Context.Render method, but without requiring
	// having an echo.Context for use. It's useful for rendering templated responses from non-echo
	// middleware, e.g. the error handler in github.com/gorilla/csrf
	buf := new(bytes.Buffer)
	if rerr := renderer.RenderWithoutContext(
		buf, templateName, NewRenderData(r, g, templateData, authData),
	); rerr != nil {
		return rerr
	}

	// Write render result
	w.Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	w.WriteHeader(status)
	_, werr := w.Write(buf.Bytes())
	return werr
}

func Render(
	c echo.Context, templateName string, templateData interface{}, authData interface{},
	te TemplateEtagSegments, g TemplateGlobals,
) error {
	type EtagInputs struct {
		Data interface{}
		Auth interface{}
	}
	if noContent, err := te.SetAndCheckEtag(c.Response(), c.Request(), templateName, EtagInputs{
		Data: templateData,
		Auth: authData,
	}); noContent || (err != nil) {
		return err
	}
	return c.Render(
		http.StatusOK, templateName, NewRenderData(c.Request(), g, templateData, authData),
	)
}

// Route Handlers

type Templated struct {
	Path         string
	Method       string
	HandlerMaker func(
		tg TemplateGlobals, te TemplateEtagSegments,
	) (echo.HandlerFunc, error)
	Templates []string
}

func (route Templated) AssembleTemplateEtagSegments(
	tg TemplateGlobals,
) (TemplateEtagSegments, error) {
	segments := make(TemplateEtagSegments)
	for _, templateName := range route.Templates {
		globalSegments, err := tg.GetEtagSegments(templateName)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf(
				"couldn't get the global etag segments for template %s", templateName,
			))
		}

		segments[templateName] = globalSegments
	}
	return segments, nil
}

func RegisterTemplated(e EchoRouter, r []Templated, tg TemplateGlobals) error {
	regFuncs := GetRegistrationFuncs(e)
	for _, route := range r {
		reg, ok := regFuncs[route.Method]
		if !ok {
			return errors.Errorf("unknown route %s", route.Method)
		}

		e, err := route.AssembleTemplateEtagSegments(tg)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf(
				"couldn't assemble template etag segments for route %s", route.Path),
			)
		}

		h, err := route.HandlerMaker(tg, e)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf(
				"couldn't make the handler for templated route %s", route.Path),
			)
		}

		reg(route.Path, h)
	}
	return nil
}

func CollectTemplated(collections ...[]Templated) []Templated {
	collected := make([]Templated, 0)
	for _, collection := range collections {
		collected = append(collected, collection...)
	}
	return collected
}

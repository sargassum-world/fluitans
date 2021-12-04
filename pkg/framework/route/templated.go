package route

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// Template rendering

type Meta struct {
	Path string
}

type RenderData struct {
	Meta    Meta
	Inlines Inlines
	Data    interface{}
}

func MakeRenderData(
	c echo.Context, g TemplateGlobals, data interface{},
) RenderData {
	return RenderData{
		Meta: Meta{
			Path: c.Request().URL.Path,
		},
		Inlines: g.Inlines,
		Data:    data,
	}
}

func Render(
	c echo.Context, templateName string, templateData interface{},
	te TemplateEtagSegments, g TemplateGlobals,
) error {
	templateEtagSegment, err := te.GetSegment(templateName)
	if err != nil {
		return err
	}
	noContent, err := ProcessEtag(c, templateEtagSegment, templateData)
	if err != nil || noContent {
		return err
	}

	return c.Render(http.StatusOK, templateName, MakeRenderData(c, g, templateData))
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

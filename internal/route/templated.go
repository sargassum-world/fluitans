package route

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/fingerprint"
	"github.com/sargassum-eco/fluitans/internal/fsutil"
	"github.com/sargassum-eco/fluitans/internal/template"
)

type TemplateFingerprints struct {
	App  string
	Page map[string]string
}

func ComputeTemplateFingerprints(
	layoutFiles, pageFiles, embedAssets []string,
	templates, app fs.FS,
) (*TemplateFingerprints, error) {
	pageFingerprints, err := fingerprint.ComputeFiles(pageFiles, templates)
	if err != nil {
		return nil, err
	}

	embedConcatenated, err := fsutil.ReadConcatenated(embedAssets, app)
	if err != nil {
		return nil, err
	}

	layoutConcatenated, err := fsutil.ReadConcatenated(layoutFiles, templates)
	if err != nil {
		return nil, err
	}

	g := TemplateFingerprints{
		App:  fingerprint.Compute(append(embedConcatenated, layoutConcatenated...)),
		Page: pageFingerprints,
	}
	return &g, nil
}

type TemplateGlobals struct {
	Embeds               template.Embeds
	TemplateFingerprints TemplateFingerprints
}

func (g TemplateGlobals) GetEtagSegments(
	templateName string,
) ([]string, error) {
	appFingerprint := g.TemplateFingerprints.App
	if templateName == "" {
		return []string{appFingerprint}, nil
	}

	pageFingerprint, ok := g.TemplateFingerprints.Page[templateName]
	if !ok {
		return []string{appFingerprint}, fmt.Errorf("couldn't find template %s", templateName)
	}

	return []string{appFingerprint, pageFingerprint}, nil
}

type TemplateEtagSegments map[string][]string

func (te TemplateEtagSegments) NewNotFoundError(t string) error {
	templates := make([]string, 0, len(te))
	for template := range te {
		templates = append(templates, template)
	}
	return fmt.Errorf("couldn't find template %s in [%s]", t, strings.Join(templates, ", "))
}

type Templated struct {
	Path         string
	Method       string
	HandlerMaker func(g TemplateGlobals, te TemplateEtagSegments) (echo.HandlerFunc, error)
	Templates    []string
}

func (route Templated) AssembleTemplateEtagSegments(g TemplateGlobals) (TemplateEtagSegments, error) {
	segments := make(TemplateEtagSegments)
	for _, templateName := range route.Templates {
		globalSegments, err := g.GetEtagSegments(templateName)
		if err != nil {
			return nil, err
		}

		segments[templateName] = globalSegments
	}
	return segments, nil
}

func RegisterTemplated(e EchoRouter, r []Templated, g TemplateGlobals) error {
	regFuncs := GetRegistrationFuncs(e)
	for _, route := range r {
		reg, ok := regFuncs[route.Method]
		if !ok {
			return fmt.Errorf("unknown route %s", route.Method)
		}

		e, err := route.AssembleTemplateEtagSegments(g)
		if err != nil {
			return err
		}

		h, err := route.HandlerMaker(g, e)
		if err != nil {
			return err
		}

		reg(route.Path, h)
	}
	return nil
}

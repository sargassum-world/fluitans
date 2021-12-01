package route

import (
	"fmt"
	"io/fs"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/fingerprint"
	"github.com/sargassum-eco/fluitans/internal/fsutil"
	"github.com/sargassum-eco/fluitans/internal/template"
)

type TemplateFingerprints struct {
	App  string
	Page map[string]string
}

func ComputeTemplateFingerprints(
	layoutFiles, pageFiles, appAssets []string,
	templates, app fs.FS,
) (*TemplateFingerprints, error) {
	// TODO: instead of App having all partials & layouts everywhere, it should
	// only have the ones in shared; then the page fingerprints should only
	// depend on the partials & layouts in the same top-level subdirectory of templates
	pageFingerprints, err := fingerprint.ComputeFiles(pageFiles, templates)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't compute templated page fingerprints")
	}

	appConcatenated, err := fsutil.ReadConcatenated(appAssets, app)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't concatenate all app assets")
	}

	layoutConcatenated, err := fsutil.ReadConcatenated(layoutFiles, templates)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't concatenate all layout & partials templates")
	}

	tf := TemplateFingerprints{
		App:  fingerprint.Compute(append(appConcatenated, layoutConcatenated...)),
		Page: pageFingerprints,
	}
	return &tf, nil
}

type TemplateGlobals struct {
	Embeds               template.Embeds
	TemplateFingerprints TemplateFingerprints
	App                  interface{}
}

func (tg TemplateGlobals) GetEtagSegments(
	templateName string,
) ([]string, error) {
	appFingerprint := tg.TemplateFingerprints.App
	if templateName == "" {
		return []string{appFingerprint}, nil
	}

	pageFingerprint, ok := tg.TemplateFingerprints.Page[templateName]
	if !ok {
		return []string{appFingerprint}, errors.Errorf(
			"couldn't find page fingerprint for template %s", templateName,
		)
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
	HandlerMaker func(tg TemplateGlobals, te TemplateEtagSegments) (echo.HandlerFunc, error)
	Templates    []string
}

func (route Templated) AssembleTemplateEtagSegments(tg TemplateGlobals) (TemplateEtagSegments, error) {
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

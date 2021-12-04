package route

import (
	"io/fs"

	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework/fingerprint"
	"github.com/sargassum-eco/fluitans/pkg/framework/fsutil"
	"github.com/sargassum-eco/fluitans/pkg/framework/template"
)

type Inlines struct {
	CSS template.EmbeddedCSSAssets
	JS  template.EmbeddedJSAssets
}

func NewInlines(css map[string]string, js map[string]string) Inlines {
	return Inlines{
		CSS: template.PreprocessCSS(css),
		JS:  template.PreprocessJS(js),
	}
}

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
	Inlines              Inlines
	TemplateFingerprints TemplateFingerprints
	App                  interface{}
}

func (tg TemplateGlobals) GetEtagSegments(templateName string) ([]string, error) {
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

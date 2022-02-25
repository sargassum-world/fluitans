package route

import (
	"fmt"
	"io/fs"
	"path"

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

func ComputeAppFingerprint(
	appAssets, sharedTemplates []string, templates, app fs.FS,
) (string, error) {
	appConcatenated, err := fsutil.ReadConcatenated(appAssets, app)
	if err != nil {
		return "", errors.Wrap(err, "couldn't load all app assets together for fingerprinting")
	}
	sharedConcatenated, err := fsutil.ReadConcatenated(sharedTemplates, templates)
	if err != nil {
		return "", errors.Wrap(err, "couldn't load all shared templates together for fingerprinting")
	}

	return fingerprint.Compute(append(appConcatenated, sharedConcatenated...)), nil
}

func ComputePageFingerprints(
	moduleNonpageFiles map[string][]string, pageFiles []string, templates fs.FS,
) (map[string]string, error) {
	moduleNonpages := make(map[string][]byte)
	for module, files := range moduleNonpageFiles {
		loadedNonpages, err := fsutil.ReadConcatenated(files, templates)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf(
				"couldn't load non-page template files in template module %s for fingerprinting", module,
			))
		}
		moduleNonpages[module] = loadedNonpages
	}

	pages := make(map[string][]byte)
	for _, file := range pageFiles {
		loadedPage, err := fsutil.ReadFile(file, templates)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf(
				"couldn't load page template %s for fingerprinting", file,
			))
		}
		pages[file] = loadedPage
	}

	pageFingerprints := make(map[string]string)
	for _, pageFile := range pageFiles {
		module := path.Dir(pageFile)
		pageFingerprints[pageFile] = fingerprint.Compute(append(
			// Each page's fingerprint is computed from the page template itself as well as any non-page
			// files (e.g. partials) within its module, recursively including all non-page files in
			// submodules (i.e. subdirectories)
			pages[pageFile], moduleNonpages[module]...,
		))
	}
	return pageFingerprints, nil
}

type TemplateGlobals struct {
	Inlines          Inlines
	AppFingerprint   string
	PageFingerprints map[string]string
}

func (tg TemplateGlobals) GetEtagSegments(templateName string) ([]string, error) {
	appFingerprint := tg.AppFingerprint
	if templateName == "" {
		return []string{appFingerprint}, nil
	}

	pageFingerprint, ok := tg.PageFingerprints[templateName]
	if !ok {
		return []string{appFingerprint}, errors.Errorf(
			"couldn't find page fingerprint for template %s", templateName,
		)
	}

	return []string{appFingerprint, pageFingerprint}, nil
}

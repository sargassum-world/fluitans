// Package embeds provides support for embedding all web app files
package embeds

import (
	"fmt"
	"html/template"
	"io/fs"

	"github.com/benbjohnson/hashfs"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework/fsutil"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	tp "github.com/sargassum-eco/fluitans/pkg/framework/template"
)

type Embeds struct {
	StaticFS    fs.FS
	StaticHFS   *hashfs.FS
	TemplatesFS fs.FS
	AppFS       fs.FS
	AppHFS      *hashfs.FS
	FontsFS     fs.FS
	Inlines     route.Inlines
}

func (e Embeds) IdentifyModuleNonpageFiles() (map[string][]string, error) {
	modules, err := fsutil.ListDirectories(e.TemplatesFS, tp.FilterModule)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't list template modules")
	}

	moduleFiles := make(map[string][]string)
	for _, module := range modules {
		var subfs fs.FS
		if module == "" {
			subfs = e.TemplatesFS
		} else {
			subfs, err = fs.Sub(e.TemplatesFS, module)
		}
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("couldn't list template module %s", module))
		}
		moduleSubfiles, err := fsutil.ListFiles(subfs, tp.FilterNonpage)
		moduleFiles[module] = make([]string, len(moduleSubfiles))
		for i, subfile := range moduleSubfiles {
			if module == "" {
				moduleFiles[module][i] = subfile
			} else {
				moduleFiles[module][i] = fmt.Sprintf("%s/%s", module, subfile)
			}
		}
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("couldn't list template files in module %s", module))
		}
	}
	return moduleFiles, nil
}

func (e Embeds) ComputeTemplateFingerprints() (*route.TemplateFingerprints, error) {
	sharedFiles, err := fsutil.ListFiles(e.TemplatesFS, tp.FilterShared)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't list shared templates")
	}

	moduleNonpageFiles, err := e.IdentifyModuleNonpageFiles()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't list template module non-page template files")
	}

	pageFiles, err := fsutil.ListFiles(e.TemplatesFS, tp.FilterPage)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't list template pages")
	}

	appAssetFiles, err := fsutil.ListFiles(e.AppFS, tp.FilterAsset)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't list app assets")
	}

	return route.ComputeTemplateFingerprints(
		sharedFiles, moduleNonpageFiles, pageFiles, appAssetFiles, e.TemplatesFS, e.AppFS,
	)
}

func (e Embeds) MakeStaticGlobals() route.StaticGlobals {
	return route.StaticGlobals{
		FS: map[string]fs.FS{
			"Web":   e.StaticFS,
			"Fonts": e.FontsFS,
		},
		HFS: map[string]fs.FS{
			"Static": e.StaticHFS,
			"App":    e.AppHFS,
		},
	}
}

func (e Embeds) NewTemplateRenderer(functions ...template.FuncMap) *tp.TemplateRenderer {
	return tp.NewTemplateRenderer(
		e.TemplatesFS, append(functions, tp.FuncMap(e.AppHFS.HashName, e.StaticHFS.HashName))...,
	)
}

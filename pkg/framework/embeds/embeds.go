// Package embeds provides support for embedding all web app files
package embeds

import (
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

func (e Embeds) ComputeAppFingerprint() (fingerprint string, err error) {
	appAssetFiles, err := fsutil.ListFiles(e.AppFS, tp.FilterAsset)
	if err != nil {
		err = errors.Wrap(err, "couldn't list app assets")
		return
	}
	sharedFiles, err := fsutil.ListFiles(e.TemplatesFS, tp.FilterShared)
	if err != nil {
		err = errors.Wrap(err, "couldn't list shared templates")
		return
	}
	fingerprint, err = route.ComputeAppFingerprint(appAssetFiles, sharedFiles, e.TemplatesFS, e.AppFS)
	if err != nil {
		err = errors.Wrap(err, "couldn't compute fingerprint for app")
		return
	}
	return
}

func (e Embeds) ComputePageFingerprints() (fingerprints map[string]string, err error) {
	moduleNonpageFiles, err := IdentifyModuleNonpageFiles(e.TemplatesFS)
	if err != nil {
		err = errors.Wrap(err, "couldn't list template module non-page template files")
		return
	}
	pageFiles, err := fsutil.ListFiles(e.TemplatesFS, tp.FilterPage)
	if err != nil {
		err = errors.Wrap(err, "couldn't list template pages")
		return
	}

	fingerprints, err = route.ComputePageFingerprints(moduleNonpageFiles, pageFiles, e.TemplatesFS)
	if err != nil {
		err = errors.Wrap(err, "couldn't compute fingerprint for page/module templates")
		return
	}
	return
}

func (e Embeds) NewTemplateRenderer(functions ...template.FuncMap) *tp.TemplateRenderer {
	return tp.NewTemplateRenderer(
		e.TemplatesFS, append(functions, tp.FuncMap(e.AppHFS.HashName, e.StaticHFS.HashName))...,
	)
}

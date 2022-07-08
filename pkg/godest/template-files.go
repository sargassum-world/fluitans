package godest

import (
	"html/template"
	"io/fs"
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
		return nil, errors.Errorf("no files named for parsing")
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
		b, err = fs.ReadFile(fsys, file)
		return file, b, err
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
			return nil, errors.Errorf("pattern matches no files: %#q", pattern)
		}
		filenames = append(filenames, list...)
	}
	return parseFiles(t, readFileFS(fsys), filenames...)
}

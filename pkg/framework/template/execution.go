// Package template preloads a collection of shared template files and page template files for fast
// execution of page templates, and specifies a layout convention to enable this functionality.
package template

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
)

type Templates struct {
	allTemplates  *template.Template
	pageTemplates map[string]*template.Template
	templatesFS   fs.FS
}

func NewTemplates(fsys fs.FS, pageFiles []string, functions ...template.FuncMap) Templates {
	tmpl := template.New("App")
	for _, f := range functions {
		tmpl = tmpl.Funcs(f)
	}
	allTemplates := template.Must(ParseFS(tmpl, fsys, "**/*"+FileExt))
	pageTemplates := make(map[string]*template.Template)
	for _, name := range pageFiles {
		all := template.Must(allTemplates.Clone())
		pageTemplates[name] = template.Must(ParseFS(all, fsys, name))
	}
	return Templates{
		allTemplates:  allTemplates,
		pageTemplates: pageTemplates,
		templatesFS:   fsys,
	}
}

func (t *Templates) Execute(w io.Writer, name string, data interface{}) error {
	tmpl, ok := t.pageTemplates[name]
	if !ok {
		return fmt.Errorf("template %s not found", name)
	}
	return tmpl.ExecuteTemplate(w, name, data)
}

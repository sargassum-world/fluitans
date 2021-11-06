package template

import (
	"fmt"
	"html/template"
	"io/fs"
)

// parseFiles is a direct copy of the parseFiles helper function in text/template.
func parseFiles(
	t *template.Template,
	readFile func(string) (string, []byte, error),
	filenames ...string,
) (*template.Template, error) {
	if len(filenames) == 0 {
		return nil, fmt.Errorf("template: no files named in call to ParseFiles")
	}

	for _, filename := range filenames {
		name, b, err := readFile(filename)
		if err != nil {
			return nil, err
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
			return nil, err
		}
	}
	return t, nil
}

// readFileFS is an adaptation of the readFileFS helper function in text/template,
// except it uses the fully-qualified path name of the template file in the filesystem
// rather than just the template file's basename.
func readFileFS(fsys fs.FS) func(string) (string, []byte, error) {
	return func(file string) (name string, b []byte, err error) {
		name = file
		b, err = fs.ReadFile(fsys, file)
		return
	}
}

// ParseFS is a direct copy of the parseFS helper function in text/template,
// except it uses the fully-qualified path name of the template file in the filesystem
// rather than just the template file's basename.
func ParseFS(t *template.Template, fsys fs.FS, patterns ...string) (*template.Template, error) {
	var filenames []string
	for _, pattern := range patterns {
		list, err := fs.Glob(fsys, pattern)
		if err != nil {
			return nil, err
		}
		if len(list) == 0 {
			return nil, fmt.Errorf("template: pattern matches no files: %#q", pattern)
		}
		filenames = append(filenames, list...)
	}
	return parseFiles(t, readFileFS(fsys), filenames...)
}

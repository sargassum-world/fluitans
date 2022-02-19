package template

import (
	"fmt"
	"html/template"
)

// All functions

func FuncMap(appNamer, staticNamer HashNamer) template.FuncMap {
	return template.FuncMap{
		"appHashed":     getHashedName("app", appNamer),
		"staticHashed":  getHashedName("static", staticNamer),
	}
}

// Asset hashed naming

type HashNamer func(string) string

func getHashedName(root string, namer HashNamer) HashNamer {
	return func(file string) string {
		return fmt.Sprintf("/%s/%s", root, namer(file))
	}
}

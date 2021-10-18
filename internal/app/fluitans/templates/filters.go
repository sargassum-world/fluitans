package templates

import (
	"strings"
)

func Filter(path string) bool {
	return strings.HasSuffix(path, ".tmpl")
}

func FilterApp(path string) bool {
	return strings.HasSuffix(path, ".layout.tmpl") || strings.HasSuffix(path, ".partial.tmpl")
}

func FilterPage(path string) bool {
	return Filter(path) && !FilterApp(path)
}

package template

import (
	"strings"
)

// Template Files

func Filter(path string) bool {
	return strings.HasSuffix(path, ".tmpl")
}

func FilterApp(path string) bool {
	return strings.HasSuffix(path, ".layout.tmpl") || strings.HasSuffix(path, ".partial.tmpl")
}

func FilterPage(path string) bool {
	return Filter(path) && !FilterApp(path)
}

// Built asset Files

func FilterAsset(path string) bool {
	return strings.HasSuffix(path, ".min.css") || strings.HasSuffix(path, ".js")
}

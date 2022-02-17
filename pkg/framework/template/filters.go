package template

import (
	"strings"
)

// Template Files

func Filter(path string) bool {
	return strings.HasSuffix(path, ".tmpl")
}

func FilterShared(path string) bool {
	return Filter(path) && strings.HasPrefix(path, "shared/")
}

func FilterPartial(path string) bool {
	return strings.HasSuffix(path, ".partial.tmpl")
}

func FilterLayout(path string) bool {
	return strings.HasSuffix(path, ".layout.tmpl")
}

func FilterNonpage(path string) bool {
	return FilterPartial(path) || FilterLayout(path)
}

func FilterPage(path string) bool {
	// Usually a page ends with ".page.tmpl", but it may end with other things,
	// e.g. ".webmanifest.tmpl", ".json.tmpl", etc.
	return Filter(path) && !FilterNonpage(path)
}

func FilterModule(path string) bool {
	return path != "shared"
}

// Built asset Files

func FilterAsset(path string) bool {
	return strings.HasSuffix(path, ".min.css") || strings.HasSuffix(path, ".js")
}

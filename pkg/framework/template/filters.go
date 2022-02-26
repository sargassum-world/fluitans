package template

import (
	"strings"
)

// Template Files

const (
	FileExt        = ".tmpl"
	PartialFileExt = ".partial" + FileExt
	LayoutFileExt  = ".layout" + FileExt
	SharedModule   = "shared"
)

func Filter(path string) bool {
	return strings.HasSuffix(path, FileExt)
}

func FilterShared(path string) bool {
	return Filter(path) && strings.HasPrefix(path, SharedModule+"/")
}

func FilterPartial(path string) bool {
	return strings.HasSuffix(path, PartialFileExt)
}

func FilterLayout(path string) bool {
	return strings.HasSuffix(path, LayoutFileExt)
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
	return path != SharedModule
}

// Built Asset Files

func FilterAsset(path string) bool {
	return strings.HasSuffix(path, ".min.css") || strings.HasSuffix(path, ".js")
}

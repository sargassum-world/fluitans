package templates

import (
	"html/template"
)

type Meta struct {
	Description string
	Path        string
}

type EmbedAssets struct {
	BundleEagerCSS template.CSS
	BundleEagerJS  template.JS
}

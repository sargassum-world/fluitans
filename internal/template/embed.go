package template

import (
	"html/template"
)

type (
	EmbeddableAssets  map[string]string
	EmbeddedCSSAssets map[string]template.CSS
	EmbeddedJSAssets  map[string]template.JS
)

type Embeds struct {
	CSS EmbeddedCSSAssets
	JS  EmbeddedJSAssets
}

func PreprocessCSS(assets EmbeddableAssets) EmbeddedCSSAssets {
	e := make(EmbeddedCSSAssets)
	for key, value := range assets {
		e[key] = template.CSS(value)
	}
	return e
}

func PreprocessJS(assets EmbeddableAssets) EmbeddedJSAssets {
	e := make(EmbeddedJSAssets)
	for key, value := range assets {
		//nolint:gosec // This bundle is generated from code in web/app/src, so we know it's well-formed.
		e[key] = template.JS(value)
	}
	return e
}

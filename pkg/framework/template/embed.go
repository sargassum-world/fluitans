// Package template structures the use of HTML templates to render pages
package template

import (
	"html/template"
)

type (
	EmbeddedCSSAssets map[string]template.CSS
	EmbeddedJSAssets  map[string]template.JS
)

func PreprocessCSS(assets map[string]string) EmbeddedCSSAssets {
	e := make(EmbeddedCSSAssets)
	for key, value := range assets {
		e[key] = template.CSS(value)
	}
	return e
}

func PreprocessJS(assets map[string]string) EmbeddedJSAssets {
	e := make(EmbeddedJSAssets)
	for key, value := range assets {
		//nolint:gosec // This is generated from code in web/app/src, so we know it's well-formed.
		e[key] = template.JS(value)
	}
	return e
}

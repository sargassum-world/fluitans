// Package web contains web application-specific components: static web assets,
// server-side templates, and modest JS sprinkles and spots.
package web

import (
	"embed"
	"io/fs"

	"github.com/benbjohnson/hashfs"
)

//go:embed static/*
var staticFS embed.FS
var StaticFS, _ = fs.Sub(staticFS, "static")
var StaticHFS = hashfs.NewFS(StaticFS)

//go:embed templates/*.tmpl templates/*/*.tmpl
var templatesFS embed.FS
var TemplatesFS, _ = fs.Sub(templatesFS, "templates")

//go:embed app/public/build/* app/public/build/fonts/*
var appFS embed.FS
var AppFS, _ = fs.Sub(appFS, "app/public/build")
var AppHFS = hashfs.NewFS(AppFS)

//go:embed app/public/build/bundle-eager.js
var BundleEagerJS string

//go:embed app/public/build/theme-eager.min.css
var BundleEagerCSS string

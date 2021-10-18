// Package web contains web application-specific components: static web assets,
// server-side templates, and modest JS sprinkles and spots.
package web

import (
	"embed"
	"io/fs"

	"github.com/benbjohnson/hashfs"
)

var (
	//go:embed static/*
	staticFS    embed.FS
	StaticFS, _ = fs.Sub(staticFS, "static")
	StaticHFS   = hashfs.NewFS(StaticFS)
)

var (
	//go:embed templates/*.tmpl templates/*/*.tmpl
	templatesFS    embed.FS
	TemplatesFS, _ = fs.Sub(templatesFS, "templates")
)

var (
	//go:embed app/public/build/*
	appFS    embed.FS
	AppFS, _ = fs.Sub(appFS, "app/public/build")
	AppHFS   = hashfs.NewFS(AppFS)
)

var (
	//go:embed app/public/build/fonts/*
	fontsFS    embed.FS
	FontsFS, _ = fs.Sub(fontsFS, "app/public/build/fonts")
)

//go:embed app/public/build/bundle-eager.js
var BundleEagerJS string

//go:embed app/public/build/theme-eager.min.css
var BundleEagerCSS string

// Package web contains web application-specific components: static web assets,
// server-side templates, and modest JS sprinkles and spots.
package web

import (
	"embed"
	"io/fs"

	"github.com/benbjohnson/hashfs"

	"github.com/sargassum-eco/fluitans/pkg/framework/embeds"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

var (
	//go:embed static/*
	staticEFS   embed.FS
	staticFS, _ = fs.Sub(staticEFS, "static")
	staticHFS   = hashfs.NewFS(staticFS)
)

var (
	//go:embed templates/*
	templatesEFS   embed.FS
	templatesFS, _ = fs.Sub(templatesEFS, "templates")
)

var (
	//go:embed app/public/build/*
	appEFS   embed.FS
	appFS, _ = fs.Sub(appEFS, "app/public/build")
	appHFS   = hashfs.NewFS(appFS)
)

var (
	//go:embed app/public/build/fonts/*
	fontsEFS   embed.FS
	fontsFS, _ = fs.Sub(fontsEFS, "app/public/build/fonts")
)

//go:embed app/public/build/bundle-eager.js
var bundleEagerJS string

//go:embed app/public/build/theme-eager.min.css
var bundleEagerCSS string

func NewEmbeds() embeds.Embeds {
	return embeds.Embeds{
		StaticFS:    staticFS,
		StaticHFS:   staticHFS,
		TemplatesFS: templatesFS,
		AppFS:       appFS,
		AppHFS:      appHFS,
		FontsFS:     fontsFS,
		Inlines: route.NewInlines(
			map[string]string{
				"BundleEager": bundleEagerCSS,
			},
			map[string]string{
				"BundleEager": bundleEagerJS,
			},
		),
	}
}

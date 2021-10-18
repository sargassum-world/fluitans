package fluitans

import (
	"github.com/sargassum-eco/fluitans/internal/template"
	"github.com/sargassum-eco/fluitans/web"
)

var embeds = template.Embeds{
	CSS: template.PreprocessCSS(template.EmbeddableAssets{
		"BundleEager": web.BundleEagerCSS,
	}),
	JS: template.PreprocessJS(template.EmbeddableAssets{
		"BundleEager": web.BundleEagerJS,
	}),
}

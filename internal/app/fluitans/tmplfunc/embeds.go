package tmplfunc

import (
	"github.com/sargassum-eco/fluitans/pkg/framework"
)

type HashedNamers struct {
	AppHashed func(filename string) string
	StaticHashed func(filename string) string
}

func NewHashedNamers(appURLPrefix, staticURLPrefix string, embeds framework.Embeds) HashedNamers {
	return HashedNamers{
		AppHashed: embeds.GetAppHashedNamer(appURLPrefix),
		StaticHashed: embeds.GetStaticHashedNamer(staticURLPrefix),
	}
}

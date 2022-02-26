package godest

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/benbjohnson/hashfs"

	"github.com/sargassum-eco/fluitans/pkg/godest/httpcache"
)

func HandleFS(routePrefix string, fsys fs.FS, age time.Duration) http.Handler {
	return httpcache.WrapStaticHeader(
		http.StripPrefix(routePrefix, http.FileServer(http.FS(fsys))), int(age.Seconds()),
	)
}

func HandleFSFileRevved(routePrefix string, fsys fs.FS) http.Handler {
	return http.StripPrefix(routePrefix, hashfs.FileServer(fsys))
}

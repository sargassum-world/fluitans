package godest

import (
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/benbjohnson/hashfs"
)

func wrapStaticHeader(h http.Handler, age int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", age))
		h.ServeHTTP(w, r)
	})
}

func HandleFS(routePrefix string, fsys fs.FS, age time.Duration) http.Handler {
	return wrapStaticHeader(
		http.StripPrefix(routePrefix, http.FileServer(http.FS(fsys))), int(age.Seconds()),
	)
}

func HandleFSFileRevved(routePrefix string, fsys fs.FS) http.Handler {
	return http.StripPrefix(routePrefix, hashfs.FileServer(fsys))
}

package godest

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/benbjohnson/hashfs"
	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/pkg/godest/httpcache"
)

func HandleFS(routePrefix string, fsys fs.FS, age time.Duration) echo.HandlerFunc {
	return echo.WrapHandler(httpcache.WrapStaticHeader(
		http.StripPrefix(routePrefix, http.FileServer(http.FS(fsys))), int(age.Seconds()),
	))
}

func HandleFSFileRevved(routePrefix string, fsys fs.FS) echo.HandlerFunc {
	return echo.WrapHandler(http.StripPrefix(routePrefix, hashfs.FileServer(fsys)))
}

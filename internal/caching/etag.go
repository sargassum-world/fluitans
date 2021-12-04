package caching

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// Etag Assembly

const etagSegmentDelimiter = " "

func JoinEtagSegments(segments ...string) string {
	return strings.Join(segments, etagSegmentDelimiter)
}

func MakeEtag(segments ...string) string {
	return fmt.Sprintf("W/\"%s\"", strings.Join(segments, etagSegmentDelimiter))
}

// Headers for Etags

func SetNoEtag(resh http.Header) {
	resh.Set("Cache-Control", "no-store, max-age=0")
}

func SetEtag(resh http.Header, etag string) {
	resh.Set("Cache-Control", "no-cache")
	resh.Set("Etag", etag)
}

func CheckEtagMatch(reqh http.Header, etag string) bool {
	match := reqh.Get("If-None-Match")
	if match == "" || etag == "" {
		return false
	}

	return match == etag
}

func ProcessEtag(c echo.Context, etagSegments []string) (bool, error) {
	etag := MakeEtag(JoinEtagSegments(etagSegments...))
	SetEtag(c.Response().Header(), etag)
	if CheckEtagMatch(c.Request().Header, etag) {
		return true, c.NoContent(http.StatusNotModified)
	}

	return false, nil
}

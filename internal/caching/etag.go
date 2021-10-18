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

func SetNoEtag(c echo.Context) {
	c.Response().Header().Set("Cache-Control", "no-store, max-age=0")
}

func SetEtag(c echo.Context, etag string) {
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Etag", etag)
}

func CheckEtagMatch(c echo.Context, etag string) bool {
	match := c.Request().Header.Get("If-None-Match")
	if match == "" || etag == "" {
		return false
	}

	return match == etag
}

func ProcessEtag(c echo.Context, templateEtagSegments []string, dataEtagSegments ...string) (bool, error) {
	etag := MakeEtag(JoinEtagSegments(
		append(templateEtagSegments, dataEtagSegments...)...,
	))
	SetEtag(c, etag)
	if CheckEtagMatch(c, etag) {
		return true, c.NoContent(http.StatusNotModified)
	}

	return false, nil
}

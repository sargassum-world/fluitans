package httpcache

import (
	"fmt"
	"net/http"
	"strings"
)

// Etag Assembly

const etagSegmentDelimiter = " "

func JoinEtagSegments(segments ...string) string {
	return strings.Join(segments, etagSegmentDelimiter)
}

func MakeEtag(segments ...string) string {
	return fmt.Sprintf("W/\"%s\"", JoinEtagSegments(segments...))
}

// Headers for Etags

func SetNoEtag(resh http.Header) {
	// Don't cache cookies - see the "Web Content Caching" subsection of
	// https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
	resh.Set("Cache-Control", "no-cache=\"Set-Cookie, Set-Cookie2\", no-store, max-age=0")
}

func SetEtag(resh http.Header, etag string) {
	resh.Set("Cache-Control", "private, no-cache")
	resh.Set("Etag", etag)
}

func CheckEtagMatch(reqh http.Header, etag string) bool {
	match := reqh.Get("If-None-Match")
	if match == "" || etag == "" {
		return false
	}
	return match == etag
}

func SetAndCheckEtag(w http.ResponseWriter, r *http.Request, etagSegments ...string) bool {
	etag := MakeEtag(etagSegments...)
	SetEtag(w.Header(), etag)
	if !CheckEtagMatch(r.Header, etag) {
		return false
	}
	w.WriteHeader(http.StatusNotModified)
	return true
}

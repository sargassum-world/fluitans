package godest

import (
	"fmt"
	"net/http"
	"strings"
)

// Etag Assembly

const etagSegmentDelimiter = " "

func makeEtag(segments ...string) string {
	return fmt.Sprintf("W/\"%s\"", strings.Join(segments, etagSegmentDelimiter))
}

// Headers for Etags

func SetUncacheable(resh http.Header) {
	// Don't cache cookies - see the "Web Content Caching" subsection of
	// https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
	resh.Set("Cache-Control", "no-cache=\"Set-Cookie, Set-Cookie2\", no-store, max-age=0")
}

func setEtag(resh http.Header, etag string) {
	resh.Set("Cache-Control", "private, no-cache")
	resh.Set("Etag", etag)
}

func checkEtagMatch(reqh http.Header, etag string) (matches bool) {
	matchQuery := reqh.Get("If-None-Match")
	if matchQuery == "" || etag == "" {
		return false
	}
	return matchQuery == etag
}

func setAndCheckEtag(
	w http.ResponseWriter, r *http.Request, etagSegments ...string,
) (noContent bool) {
	etag := makeEtag(etagSegments...)
	setEtag(w.Header(), etag)
	return checkEtagMatch(r.Header, etag)
}

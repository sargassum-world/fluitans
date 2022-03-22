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

func setEtag(resh http.Header, etag string) {
	WithAlwaysRevalidate()(resh)
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

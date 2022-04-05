package turbostreams

import (
	"net/http"
	"strings"
)

const ContentType = "text/vnd.turbo-stream.html"

func Accepted(h http.Header) bool {
	for _, a := range strings.Split(h.Get("Accept"), ",") {
		if strings.TrimSpace(a) == ContentType {
			return true
		}
	}
	return false
}

// Package turbo provides server-side support for the Hotwire Turbo library.
package turbo

import (
	"net/http"
	"strings"
)

type StreamAction string

const (
	StreamAppend  StreamAction = "append"
	StreamPrepend StreamAction = "prepend"
	StreamReplace StreamAction = "replace"
	StreamUpdate  StreamAction = "update"
	StreamRemove  StreamAction = "remove"
	StreamBefore  StreamAction = "before"
	StreamAfter   StreamAction = "after"
)

type Stream struct {
	Action   StreamAction
	Target   string
	Targets  string
	Template string
	Data     interface{}
}

const StreamContentType = "text/vnd.turbo-stream.html"

func StreamAccepted(h http.Header) bool {
	for _, a := range strings.Split(h.Get("Accept"), ",") {
		if strings.TrimSpace(a) == StreamContentType {
			return true
		}
	}
	return false
}

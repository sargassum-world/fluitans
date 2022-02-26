// Package httpcache provides utilities for working with HTTP caching in Echo
package httpcache

import (
	"fmt"
	"net/http"
)

func WrapStaticHeader(h http.Handler, age int) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", age))
		h.ServeHTTP(w, r)
	})
}

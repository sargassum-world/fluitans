// Package middleware provides middleware functionality not provided by other modules
package middleware

import (
	"net/http"
)

const (
	corpHeader = "Cross-Origin-Resource-Policy"
	coepHeader = "Cross-Origin-Embedder-Policy"
)

func SetCORP(policy string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(corpHeader, policy)
			h.ServeHTTP(w, r)
		})
	}
}

func SetCOEP(policy string) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set(coepHeader, policy)
			h.ServeHTTP(w, r)
		})
	}
}

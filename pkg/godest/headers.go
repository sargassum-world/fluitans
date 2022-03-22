package godest

import (
	"fmt"
	"net/http"
)

type HeaderOption func(http.Header)

func WithContentType(contentType string) HeaderOption {
	return func(h http.Header) {
		h.Set("Content-Type", contentType)
	}
}

func WithUncacheable() HeaderOption {
	return func(h http.Header) {
		// Don't cache cookies - see the "Web Content Caching" subsection of
		// https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
		h.Set("Cache-Control", "no-cache=\"Set-Cookie, Set-Cookie2\", no-store, max-age=0")
	}
}

func WithAlwaysRevalidate() HeaderOption {
	return func(h http.Header) {
		h.Set("Cache-Control", "private, no-cache")
	}
}

func WithRevalidateWhenStale(maxAge int) HeaderOption {
	return func(h http.Header) {
		// Don't cache cookies - see the "Web Content Caching" subsection of
		// https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
		h.Set("Cache-Control", fmt.Sprintf(
			"private, no-cache=\"Set-Cookie, Set-Cookie2\", max-age=%d, must-revalidate", maxAge,
		))
	}
}

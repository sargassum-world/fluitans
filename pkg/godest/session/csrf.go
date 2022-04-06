package session

import (
	"net/http"

	"github.com/gorilla/csrf"
)

func NewCSRFMiddleware(config Config, opts ...csrf.Option) func(http.Handler) http.Handler {
	sameSite := csrf.SameSiteDefaultMode
	switch config.CookieOptions.SameSite {
	case http.SameSiteLaxMode:
		sameSite = csrf.SameSiteLaxMode
	case http.SameSiteStrictMode:
		sameSite = csrf.SameSiteStrictMode
	case http.SameSiteNoneMode:
		sameSite = csrf.SameSiteNoneMode
	}
	options := []csrf.Option{
		csrf.Path(config.CookieOptions.Path),
		csrf.Domain(config.CookieOptions.Domain),
		csrf.MaxAge(config.CookieOptions.MaxAge),
		csrf.Secure(config.CookieOptions.Secure),
		csrf.HttpOnly(config.CookieOptions.HttpOnly),
		csrf.SameSite(sameSite),
		csrf.RequestHeader(config.CSRFOptions.HeaderName),
		csrf.FieldName(config.CSRFOptions.FieldName),
	}
	options = append(options, opts...)
	return csrf.Protect(config.AuthKey, options...)
}

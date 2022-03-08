// Package session standardizes session management with Gorilla sessions
package session

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/quasoft/memstore"
)

type Client struct {
	Config Config
	// TODO: allow configuration to use sqlite for a persistent session store
	Store sessions.Store
}

func NewMemStoreClient(c Config) *Client {
	store := memstore.NewMemStore(c.AuthKey)
	store.Options = &c.CookieOptions

	return &Client{
		Config: c,
		Store:  store,
	}
}

func (sc *Client) Get(r *http.Request) (*sessions.Session, error) {
	return sc.Store.Get(r, sc.Config.CookieName)
}

func (sc *Client) NewCSRFMiddleware(opts ...csrf.Option) func(http.Handler) http.Handler {
	sameSite := csrf.SameSiteDefaultMode
	switch sc.Config.CookieOptions.SameSite {
	case http.SameSiteLaxMode:
		sameSite = csrf.SameSiteLaxMode
	case http.SameSiteStrictMode:
		sameSite = csrf.SameSiteStrictMode
	case http.SameSiteNoneMode:
		sameSite = csrf.SameSiteNoneMode
	}
	options := []csrf.Option{
		csrf.Path(sc.Config.CookieOptions.Path),
		csrf.Domain(sc.Config.CookieOptions.Domain),
		csrf.MaxAge(sc.Config.CookieOptions.MaxAge),
		csrf.Secure(sc.Config.CookieOptions.Secure),
		csrf.HttpOnly(sc.Config.CookieOptions.HttpOnly),
		csrf.SameSite(sameSite),
		csrf.RequestHeader(sc.Config.CSRFOptions.HeaderName),
		csrf.FieldName(sc.Config.CSRFOptions.FieldName),
	}
	options = append(options, opts...)
	return csrf.Protect(sc.Config.AuthKey, options...)
}

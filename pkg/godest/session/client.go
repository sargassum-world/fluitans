package session

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
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

func (sc *Client) New(r *http.Request) (sess *sessions.Session, err error) {
	sess, err = sc.Store.New(r, sc.Config.CookieName)
	return sess, errors.Wrap(err, "couldn't make session")
}

func (sc *Client) Get(r *http.Request) (sess *sessions.Session, err error) {
	sess, err = sc.Store.Get(r, sc.Config.CookieName)
	return sess, errors.Wrap(err, "couldn't get session")
}

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

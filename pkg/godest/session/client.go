package session

import (
	"context"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"github.com/quasoft/memstore"
)

type Client struct {
	Config Config
	// TODO: allow configuration to use sqlite for a persistent session store
	Store *memstore.MemStore
}

func NewMemStoreClient(c Config) *Client {
	store := memstore.NewMemStore(c.AuthKey)
	store.Options = &c.CookieOptions

	return &Client{
		Config: c,
		Store:  store,
	}
}

func (sc *Client) New(r *http.Request) (*sessions.Session, error) {
	sess, err := sc.Store.New(r, sc.Config.CookieName)
	return sess, errors.Wrap(err, "couldn't make session")
}

func (sc *Client) Get(r *http.Request) (*sessions.Session, error) {
	sess, err := sc.Store.Get(r, sc.Config.CookieName)
	return sess, errors.Wrap(err, "couldn't get session from request")
}

func (sc *Client) Lookup(id string) (*sessions.Session, error) {
	r, err := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't generate HTTP request to get session")
	}
	encrypted, err := securecookie.EncodeMulti(sc.Config.CookieName, id, sc.Store.Codecs...)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't generate encoded HTTP cookie to get session")
	}
	r.AddCookie(sessions.NewCookie(sc.Config.CookieName, encrypted, &sc.Config.CookieOptions))
	sess, err := sc.Store.Get(r, sc.Config.CookieName)
	return sess, errors.Wrap(err, "couldn't get session without request")
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

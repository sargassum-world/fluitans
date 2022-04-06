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

type Store interface {
	CSRFOptions() CSRFOptions
	New(r *http.Request) (*sessions.Session, error)
	Get(r *http.Request) (*sessions.Session, error)
	Lookup(id string) (*sessions.Session, error)
	NewCSRFMiddleware(opts ...csrf.Option) func(http.Handler) http.Handler
}

// MemStore

type MemStore struct {
	Config Config
	Store *memstore.MemStore
}

func NewMemStore(c Config) *MemStore {
	store := memstore.NewMemStore(c.AuthKey)
	store.Options = &c.CookieOptions

	return &MemStore{
		Config: c,
		Store:  store,
	}
}

func (ss *MemStore) CSRFOptions() CSRFOptions {
	return ss.Config.CSRFOptions
}

func (ss *MemStore) New(r *http.Request) (*sessions.Session, error) {
	sess, err := ss.Store.New(r, ss.Config.CookieName)
	return sess, errors.Wrap(err, "couldn't make session")
}

func (ss *MemStore) Get(r *http.Request) (*sessions.Session, error) {
	sess, err := ss.Store.Get(r, ss.Config.CookieName)
	return sess, errors.Wrap(err, "couldn't get session from request")
}

func (ss *MemStore) Lookup(id string) (*sessions.Session, error) {
	r, err := http.NewRequestWithContext(context.Background(), "GET", "/", nil)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't generate HTTP request to get session")
	}
	encrypted, err := securecookie.EncodeMulti(ss.Config.CookieName, id, ss.Store.Codecs...)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't generate encoded HTTP cookie to get session")
	}
	r.AddCookie(sessions.NewCookie(ss.Config.CookieName, encrypted, &ss.Config.CookieOptions))
	sess, err := ss.Store.Get(r, ss.Config.CookieName)
	return sess, errors.Wrap(err, "couldn't get session without request")
}

func (ss *MemStore) NewCSRFMiddleware(opts ...csrf.Option) func(http.Handler) http.Handler {
	return NewCSRFMiddleware(ss.Config, opts...)
}

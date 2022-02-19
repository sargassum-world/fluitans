// Package sessions provides a high-level client for session management
package sessions

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/quasoft/memstore"

	"github.com/sargassum-eco/fluitans/pkg/framework/log"
	"github.com/sargassum-eco/fluitans/pkg/framework/session"
)

type Client struct {
	Config Config
	Logger log.Logger
	// TODO: allow configuration to use sqlite for a persistent session store
	Store sessions.Store
}

func (sc *Client) Get(c echo.Context) (*sessions.Session, error) {
	return session.Get(c, sc.Config.CookieName, sc.Store)
}

func (sc *Client) Regenerate(c echo.Context) (*sessions.Session, error) {
	return session.Regenerate(c, sc.Config.CookieName, sc.Store)
}

func (sc *Client) Invalidate(c echo.Context) (*sessions.Session, error) {
	return session.Invalidate(c, sc.Config.CookieName, sc.Store)
}

func NewMemStoreClient(l log.Logger) (*Client, error) {
	config, err := GetConfig()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't set up sessions client config")
	}

	store := memstore.NewMemStore(config.SessionKey)
	store.Options = &config.CookieOptions

	return &Client{
		Config: *config,
		Logger: l,
		Store:  store,
	}, nil
}

// Package sessions provides a high-level client for session management
package sessions

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/quasoft/memstore"

	"github.com/sargassum-eco/fluitans/pkg/framework/log"
)

type Client struct {
	Config Config
	Logger log.Logger
	// TODO: allow configuration to use sqlite for a persistent session store
	Store *memstore.MemStore
}

func (c *Client) Get(ctx echo.Context) (*sessions.Session, error) {
	return session.Get(c.Config.CookieName, ctx)
}

func (c *Client) Regenerate(ctx echo.Context) (*sessions.Session, error) {
	s, err := c.Get(ctx)
	if err != nil {
		return nil, err
	}

	s.ID = ""
	return s, nil
}

func (c *Client) Invalidate(ctx echo.Context) (*sessions.Session, error) {
	s, err := session.Get(c.Config.CookieName, ctx)
	if err != nil {
		return nil, err
	}

	s.Options.MaxAge = -1
	s.Values = make(map[interface{}]interface{})
	if err := s.Save(ctx.Request(), ctx.Response()); err != nil {
		return nil, err
	}
	return s, nil
}

func MakeClient(l log.Logger) (*Client, error) {
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

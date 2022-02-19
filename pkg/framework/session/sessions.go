// Package session standardizes session management with Echo and Gorilla sessions
package session

import (
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
)

func Get(ctx echo.Context, cookieName string, store sessions.Store) (*sessions.Session, error) {
	// TODO: implement idle timeout, and implement automatic renewal timeout (if we can). Refer to the
	// "Automatic Session Expiration" section of
	// https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
	// TODO: regenerate the session upon privilege change
	// TODO: log the session life cycle
	return store.Get(ctx.Request(), cookieName)
}

func Save(s *sessions.Session, ctx echo.Context) error {
	return s.Save(ctx.Request(), ctx.Response())
}

func Regenerate(
	ctx echo.Context, cookieName string, store sessions.Store,
) (*sessions.Session, error) {
	s, err := Get(ctx, cookieName, store)
	if err != nil {
		return nil, err
	}

	s.ID = ""
	return s, nil
}

func Invalidate(
	ctx echo.Context, cookieName string, store sessions.Store,
) (*sessions.Session, error) {
	s, err := Get(ctx, cookieName, store)
	if err != nil {
		return nil, err
	}

	s.Options.MaxAge = 0
	s.Values = make(map[interface{}]interface{})
	return s, nil
}

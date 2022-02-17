package auth

import (
	"encoding/gob"
	"fmt"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"

	sc "github.com/sargassum-eco/fluitans/internal/clients/sessions"
)

func GetSession(s sessions.Session) (session Session) {
	session.ID = s.ID
	return
}

func SetIdentity(s *sessions.Session, username string) {
	identity := Identity{
		Authenticated: true,
		User:          username,
	}
	s.Values["identity"] = identity
	gob.Register(identity)
}

func GetIdentity(s sessions.Session) (identity Identity, err error) {
	if s.IsNew {
		return
	}

	rawIdentity, ok := s.Values["identity"]
	if !ok {
		err = fmt.Errorf("missing field identity in session")
		return
	}
	identity, ok = rawIdentity.(Identity)
	if !ok {
		err = fmt.Errorf("unexpected type for field identity in session")
		return
	}
	return
}

func Get(s sessions.Session) (a Auth, err error) {
	a.Session = GetSession(s)
	a.Identity, err = GetIdentity(s)
	if !a.Identity.Authenticated {
		a.Session.ID = ""
	}
	return
}

func GetWithSession(ctx echo.Context, c *sc.Client) (a Auth, s *sessions.Session, err error) {
	s, err = c.Get(ctx)
	if err != nil {
		return Auth{}, nil, err
	}
	a, err = Get(*s)
	if err != nil {
		return Auth{}, s, err
	}

	return
}

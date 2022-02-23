package auth

import (
	"encoding/gob"
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"

	sessionsc "github.com/sargassum-eco/fluitans/internal/clients/sessions"
)

// Identity

func SetIdentity(s *sessions.Session, username string) {
	identity := Identity{
		Authenticated: username != "",
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
		// A zero value for Identity indicates that the session has no identity associated with it
		return
	}
	identity, ok = rawIdentity.(Identity)
	if !ok {
		err = fmt.Errorf("unexpected type for field identity in session")
		return
	}
	return
}

// CSRF

func SetCSRFBehavior(s *sessions.Session, omitToken bool) {
	behavior := CSRFBehavior{
		OmitToken: omitToken,
	}
	s.Values["csrfBehavior"] = behavior
	gob.Register(behavior)
}

func GetCSRFBehavior(s sessions.Session, sc *sessionsc.Client) (behavior CSRFBehavior, err error) {
	behavior.FieldName = sc.Config.CSRFOptions.FieldName
	if s.IsNew {
		return
	}

	rawBehavior, ok := s.Values["csrfBehavior"]
	if !ok {
		// By default, HTML responses don't omit the CSRF input fields (so they can't be cached),
		// to enable functionality with non-JS browsers
		return
	}
	behavior, ok = rawBehavior.(CSRFBehavior)
	behavior.FieldName = sc.Config.CSRFOptions.FieldName
	if !ok {
		err = fmt.Errorf("unexpected type for field csrfBehavior in session")
		return
	}
	return
}

// Access

func Get(c echo.Context, s sessions.Session, sc *sessionsc.Client) (a Auth, err error) {
	return GetFromRequest(c.Request(), s, sc)
}

func GetFromRequest(r *http.Request, s sessions.Session, sc *sessionsc.Client) (a Auth, err error) {
	a.Identity, err = GetIdentity(s)
	if err != nil {
		return
	}

	a.CSRF.Behavior, err = GetCSRFBehavior(s, sc)
	if err != nil {
		return
	}
	if !a.CSRF.Behavior.OmitToken {
		a.CSRF.Token = csrf.Token(r)
	}
	return
}

func GetWithSession(c echo.Context, sc *sessionsc.Client) (a Auth, s *sessions.Session, err error) {
	s, err = sc.Get(c)
	if err != nil {
		return Auth{}, nil, err
	}
	a, err = Get(c, *s, sc)
	if err != nil {
		return Auth{}, s, err
	}

	return
}

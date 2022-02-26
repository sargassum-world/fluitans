// Package session standardizes session management with Gorilla sessions
package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func Get(
	r *http.Request, cookieName string, store sessions.Store,
) (*sessions.Session, error) {
	// TODO: implement idle timeout, and implement automatic renewal timeout (if we can). Refer to the
	// "Automatic Session Expiration" section of
	// https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
	// TODO: regenerate the session upon privilege change
	// TODO: log the session life cycle
	return store.Get(r, cookieName)
}

func Regenerate(
	r *http.Request, cookieName string, store sessions.Store,
) (*sessions.Session, error) {
	s, err := Get(r, cookieName, store)
	if err != nil {
		return nil, err
	}

	s.ID = ""
	s.Values = make(map[interface{}]interface{})
	return s, nil
}

func Invalidate(
	r *http.Request, cookieName string, store sessions.Store,
) (*sessions.Session, error) {
	s, err := Get(r, cookieName, store)
	if err != nil {
		return nil, err
	}

	s.Options.MaxAge = 0
	s.Values = make(map[interface{}]interface{})
	return s, nil
}

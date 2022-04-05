// Package session standardizes session management with Gorilla sessions
package session

import (
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
)

// Session rotation

func Regenerate(s *sessions.Session) {
	s.ID = ""
	s.Values = make(map[interface{}]interface{})
}

func Invalidate(s *sessions.Session) {
	s.Options.MaxAge = -1
	s.Values = make(map[interface{}]interface{})
}

// Flash messages

const FlashErrorsKey = "_flash_errors"

func AddErrorMessage(s *sessions.Session, message string) {
	s.AddFlash(message, FlashErrorsKey)
}

func GetErrorMessages(s *sessions.Session) ([]string, error) {
	rawFlashes := s.Flashes(FlashErrorsKey)
	flashes := make([]string, 0, len(rawFlashes))
	for _, rawFlash := range rawFlashes {
		flash, ok := rawFlash.(string)
		if !ok {
			return nil, errors.Errorf(
				"session error message is of unexpected non-string type %T", rawFlash,
			)
		}
		flashes = append(flashes, flash)
	}
	return flashes, nil
}

// Package session provides a reusable pattern for session management
package session

import (
	"fmt"

	"github.com/gorilla/sessions"
)

// Errors

func AddErrorMessage(s *sessions.Session, message string) {
	s.AddFlash(message, "_flash_errors")
}

func GetErrorMessages(s *sessions.Session) ([]string, error) {
	rawFlashes := s.Flashes("_flash_errors")
	flashes := make([]string, 0, len(rawFlashes))
	for _, rawFlash := range rawFlashes {
		flash, ok := rawFlash.(string)
		if !ok {
			return nil, fmt.Errorf("session error message is of unexpected non-string type %T", rawFlash)
		}
		flashes = append(flashes, flash)
	}
	return flashes, nil
}

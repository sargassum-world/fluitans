// Package auth provides application-level standardization for authentication
package auth

import (
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
)

type Identity struct {
	Authenticated bool
	User          string
}

type CSRFBehavior struct {
	InlineToken bool
}

type CSRF struct {
	Config   sessions.CSRFOptions
	Behavior CSRFBehavior
	Token    string
}

type Auth struct {
	Identity Identity
	CSRF     CSRF
}

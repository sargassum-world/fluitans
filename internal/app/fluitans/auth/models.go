// Package auth provides application-level standardization for authentication
package auth

import (
	"html/template"
)

type Identity struct {
	Authenticated bool
	User          string
}

type Auth struct {
	Identity  Identity
	CSRFInput template.HTML
}

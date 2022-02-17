// Package auth provides application-level standardization for authentication
package auth

type Identity struct {
	Authenticated bool
	User          string
}

type Session struct {
	ID string
}

type Auth struct {
	Identity Identity
	Session  Session
}

// Package auth provides application-level standardization for authentication
package auth

type Identity struct {
	Authenticated bool
	User          string
}

type CSRFBehavior struct {
	OmitToken  bool
	FieldName  string
}

type CSRF struct {
	Behavior CSRFBehavior
	Token    string
}

type Auth struct {
	Identity  Identity
	CSRF      CSRF
}

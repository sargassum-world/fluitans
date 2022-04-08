package session

import (
	"crypto/subtle"
	"encoding/base64"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/gorilla/securecookie"
	"github.com/pkg/errors"
)

func NewCSRFMiddleware(config Config, opts ...csrf.Option) func(http.Handler) http.Handler {
	sameSite := csrf.SameSiteDefaultMode
	switch config.CookieOptions.SameSite {
	case http.SameSiteLaxMode:
		sameSite = csrf.SameSiteLaxMode
	case http.SameSiteStrictMode:
		sameSite = csrf.SameSiteStrictMode
	case http.SameSiteNoneMode:
		sameSite = csrf.SameSiteNoneMode
	}
	options := []csrf.Option{
		csrf.Path(config.CookieOptions.Path),
		csrf.Domain(config.CookieOptions.Domain),
		csrf.MaxAge(config.CookieOptions.MaxAge),
		csrf.Secure(config.CookieOptions.Secure),
		csrf.HttpOnly(config.CookieOptions.HttpOnly),
		csrf.SameSite(sameSite),
		csrf.RequestHeader(config.CSRFOptions.HeaderName),
		csrf.FieldName(config.CSRFOptions.FieldName),
	}
	return csrf.Protect(config.AuthKey, append(options, opts...)...)
}

// CSRF Token Checking

const tokenLength = 32 // bytes

// xorToken XORs tokens ([]byte) to provide unique-per-request CSRF tokens. It will return a masked
// token if the base token is XOR'ed with a one-time-pad. An unmaskd token will be returned if a
// masked token is XOR'ed with the one-time-pad used to mask it.
func xorToken(a, b []byte) []byte {
	// Copied from the github.com/gorilla/csrf package's xorToken function
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	res := make([]byte, n)

	for i := 0; i < n; i++ {
		res[i] = a[i] ^ b[i]
	}

	return res
}

// unmask splits the issued token (one-time-pad + masked token) and returns the unmasked request
// token for comparison.
func unmask(issued []byte) []byte {
	// Copied from the github.com/gorilla/csrf package's unmask function
	// Issued tokens are always masked and combined with the pad.
	if len(issued) != tokenLength*2 {
		return nil
	}

	// We now know the length of the byte slice.
	otp := issued[tokenLength:]
	masked := issued[:tokenLength]

	// Unmask the token by XOR'ing it against the OTP used to mask it.
	return xorToken(otp, masked)
}

// compare securely (constant-time) compares the unmasked token from the request against the real
// token from the session.
func compareTokens(a, b []byte) bool {
	// Copied from the github.com/gorilla/csrf package's compareTokens function
	// This is required as subtle.ConstantTimeCompare does not check for equal lengths in Go versions
	// prior to 1.3.
	if len(a) != len(b) {
		return false
	}

	return subtle.ConstantTimeCompare(a, b) == 1
}

// cookieStore is a signed cookie session store for CSRF tokens.
type cookieStore struct {
	name string
	sc   *securecookie.SecureCookie
}

// Get retrieves a CSRF token from the session cookie. It returns an empty token if decoding fails
// (e.g. HMAC validation fails or the named cookie doesn't exist).
func (cs *cookieStore) Get(r *http.Request) ([]byte, error) {
	// Copied from the github.com/gorilla/csrf package's cookieStore.Get method
	// Retrieve the cookie from the request
	cookie, err := r.Cookie(cs.name)
	if err != nil {
		return nil, err
	}

	token := make([]byte, tokenLength)
	// Decode the HMAC authenticated cookie.
	err = cs.sc.Decode(cs.name, cookie.Value, &token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

type CSRFTokenChecker struct {
	cs *cookieStore
}

const cookieName = "_gorilla_csrf"

func NewCSRFTokenChecker(config Config) *CSRFTokenChecker {
	sc := securecookie.New(config.AuthKey, nil)
	sc.SetSerializer(securecookie.JSONEncoder{})
	return &CSRFTokenChecker{
		cs: &cookieStore{
			name: cookieName, // Note: this assumes the default gorilla CSRF cookie name is used
			sc:   sc,
		},
	}
}

func (c *CSRFTokenChecker) Check(r *http.Request, token string) error {
	// Adapted from the github.com/csrf package's requestToken function and its csrf class's ServeHTTP
	// method.
	realToken, err := c.cs.Get(r)
	if err != nil || len(realToken) != tokenLength {
		return errors.Wrap(err, "request session lacks CSRF token")
	}

	issuedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return errors.Wrap(err, "couldn't decode issued token")
	}
	if !compareTokens(unmask(issuedToken), realToken) {
		return errors.New("issued CSRF token doesn't match session token")
	}
	return nil
}

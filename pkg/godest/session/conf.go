package session

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/env"
)

const envPrefix = "SESSIONS_"

type Timeouts struct {
	Absolute time.Duration
	// TODO: add idle timeout
	// TODO: add renewal timeout, if we can implement session renewal
}

type CSRFOptions struct {
	HeaderName string
	FieldName  string
}

type Config struct {
	AuthKey       []byte
	Timeouts      Timeouts
	CookieOptions sessions.Options
	CookieName    string
	CSRFOptions   CSRFOptions
}

func GetConfig() (c Config, err error) {
	c.AuthKey, err = getAuthKey()
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make session key config")
	}

	c.Timeouts, err = getTimeouts()
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make session timeouts config")
	}

	// TODO: when we implement idle timeout, pass that instead of absolute timeout
	c.CookieOptions, err = getCookieOptions(c.Timeouts.Absolute)
	if err != nil {
		return Config{}, errors.Wrap(err, "couldn't make cookie options config")
	}

	if c.CookieOptions.Secure {
		// The __Host- prefix requires Secure, HTTPS, no Domain, and path "/"
		c.CookieName = "__Host-Session"
	} else {
		c.CookieName = "session"
	}

	c.CSRFOptions = getCSRFOptions()
	return c, nil
}

func getAuthKey() (authKey []byte, err error) {
	authKey, err = env.GetBase64(envPrefix + "AUTH_KEY")
	if err != nil {
		return nil, err
	}

	if authKey == nil {
		authKeySize := 32
		authKey = securecookie.GenerateRandomKey(authKeySize)
		// TODO: print to the logger instead?
		fmt.Printf(
			"Record this key for future use as %sAUTH_KEY: %s\n",
			envPrefix, base64.StdEncoding.EncodeToString(authKey),
		)
	}
	return authKey, nil
}

func getTimeouts() (t Timeouts, err error) {
	const defaultAbsolute = 12 * 60 // default: 12 hours
	rawAbsolute, err := env.GetInt64(envPrefix+"TIMEOUTS_ABSOLUTE", defaultAbsolute)
	if err != nil {
		return Timeouts{}, errors.Wrap(err, "couldn't make absolute timeout config")
	}
	t.Absolute = time.Duration(rawAbsolute) * time.Minute
	return t, nil
}

func getCookieOptions(absoluteTimeout time.Duration) (o sessions.Options, err error) {
	noHTTPSOnly, err := env.GetBool(envPrefix + "COOKIE_NOHTTPSONLY")
	if err != nil {
		return sessions.Options{}, errors.Wrap(err, "couldn't make HTTPS-only config")
	}

	return sessions.Options{
		Path:     "/",
		Domain:   "",
		MaxAge:   int(absoluteTimeout.Seconds()),
		Secure:   !noHTTPSOnly,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

func getCSRFOptions() (o CSRFOptions) {
	o.HeaderName = env.GetString(envPrefix+"CSRF_HEADERNAME", "X-CSRF-Token")
	o.FieldName = env.GetString(envPrefix+"CSRF_FIELDNAME", "csrf-token")
	return o
}

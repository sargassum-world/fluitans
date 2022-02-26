package sessions

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/godest/env"
)

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

func GetConfig() (*Config, error) {
	authKey, err := getAuthKey()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make session key config")
	}

	timeouts, err := getTimeouts()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make session timeouts config")
	}

	// TODO: when we implement idle timeout, pass that instead of absolute timeout
	cookieOptions, err := getCookieOptions(timeouts.Absolute)
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make cookie options config")
	}

	var cookieName string
	if cookieOptions.Secure {
		// The __Host- prefix requires Secure, HTTPS, no Domain, and path "/"
		cookieName = "__Host-Session"
	} else {
		cookieName = "session"
	}

	csrfOptions := getCSRFOptions()

	return &Config{
		AuthKey:       authKey,
		Timeouts:      timeouts,
		CookieOptions: cookieOptions,
		CookieName:    cookieName,
		CSRFOptions:   csrfOptions,
	}, nil
}

func getAuthKey() ([]byte, error) {
	authKey, err := env.GetBase64("FLUITANS_SESSIONS_AUTH_KEY")
	if err != nil {
		return nil, err
	}

	if authKey == nil {
		authKeySize := 32
		authKey = securecookie.GenerateRandomKey(authKeySize)
		// TODO: print to the logger instead?
		fmt.Printf(
			"Record this key for future use as FLUITANS_SESSIONS_AUTH_KEY: %s\n",
			base64.StdEncoding.EncodeToString(authKey),
		)
	}

	return authKey, nil
}

func getTimeouts() (Timeouts, error) {
	var defaultAbsolute int64 = 12 * 60 // default: 12 hours
	rawAbsolute, err := env.GetInt64("FLUITANS_SESSIONS_TIMEOUTS_ABSOLUTE", defaultAbsolute)
	if err != nil {
		return Timeouts{}, errors.Wrap(err, "couldn't make absolute timeout config")
	}
	absolute := time.Duration(rawAbsolute) * time.Minute

	return Timeouts{
		Absolute: absolute,
	}, nil
}

func getCookieOptions(absoluteTimeout time.Duration) (sessions.Options, error) {
	noHTTPSOnly, err := env.GetBool("FLUITANS_SESSIONS_COOKIE_NOHTTPSONLY")
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

func getCSRFOptions() CSRFOptions {
	headerName := env.GetString("FLUITANS_SESSIONS_CSRF_HEADERNAME", "X-CSRF-Token")
	fieldName := env.GetString("FLUITANS_SESSIONS_CSRF_FIELDNAME", "csrf-token")

	return CSRFOptions{
		HeaderName: headerName,
		FieldName:  fieldName,
	}
}

package sessions

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/pkg/framework/env"
)

type Config struct {
	SessionKey    []byte
	CookieOptions sessions.Options
	CookieName    string
}

func GetConfig() (*Config, error) {
	sessionKey, err := getSessionKey()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make session key config")
	}

	cookieOptions, err := getCookieOptions()
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

	return &Config{
		SessionKey:    sessionKey,
		CookieOptions: cookieOptions,
		CookieName:    cookieName,
	}, nil
}

func getSessionKey() ([]byte, error) {
	sessionKey, err := env.GetBase64("FLUITANS_SESSIONS_AUTH_KEY")
	if err != nil {
		return nil, err
	}

	if sessionKey == nil {
		sessionKeySize := 32
		sessionKey = securecookie.GenerateRandomKey(sessionKeySize)
		// TODO: print to the logger instead?
		fmt.Printf(
			"Record this session key for future use as FLUITANS_SESSIONS_AUTH_KEY: %s\n",
			base64.StdEncoding.EncodeToString(sessionKey),
		)
	}

	return sessionKey, nil
}

func getCookieOptions() (sessions.Options, error) {
	var defaultMaxAge int64 = 12 // default: 12 hours
	rawMaxAge, err := env.GetInt64("FLUITANS_SESSIONS_COOKIE_MAXAGE", defaultMaxAge)
	maxAge := int((time.Duration(rawMaxAge) * time.Hour).Seconds())
	if err != nil {
		return sessions.Options{}, errors.Wrap(err, "couldn't make max age config")
	}

	noHTTPSOnly, err := env.GetBool("FLUITANS_SESSIONS_COOKIE_NOHTTPSONLY")
	if err != nil {
		return sessions.Options{}, errors.Wrap(err, "couldn't make HTTPS-only config")
	}

	return sessions.Options{
		Path:     "/",
		Domain:   "",
		MaxAge:   maxAge,
		Secure:   !noHTTPSOnly,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}, nil
}

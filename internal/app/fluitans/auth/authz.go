package auth

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/fluitans/internal/clients/sessions"
)

// Authorization

func (a Auth) Authorized() bool {
	// Right now there's only one user who can be authenticated, namely the admin, so this is
	// good enough for now.
	return a.Identity.Authenticated
}

func (a Auth) RequireAuthorized() error {
	if a.Authorized() {
		return nil
	}

	// We return StatusNotFound instead of StatusUnauthorized or StatusForbidden to avoid leaking
	// information about the existence of secret resources.
	// TODO: would the error message leak information? If so, we should leave it blank everywhere
	// across the app.
	return echo.NewHTTPError(http.StatusNotFound, "unauthorized")
}

func RequireAuthz(sc *sessions.Client) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			a, _, err := GetWithSession(c, sc)
			if err != nil {
				return err
			}
			if err = a.RequireAuthorized(); err != nil {
				return err
			}
			// TODO: store the auth in the request context
			return next(c)
		}
	}
}
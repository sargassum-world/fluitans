package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/pkg/framework/session"
)

type CSRFData struct {
	HeaderName string `json:"headerName,omitempty"`
	FieldName  string `json:"fieldName,omitempty"`
	Token      string `json:"token,omitempty"`
}

func (s *Service) getCSRF() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get session
		sess, err := s.sc.Get(c)
		if err != nil {
			return err
		}
		if err = session.Save(sess, c); err != nil {
			return err
		}

		// Produce output
		s.r.SetUncacheable(c.Response().Header())
		return c.JSON(http.StatusOK, CSRFData{
			HeaderName: s.sc.Config.CSRFOptions.HeaderName,
			FieldName:  s.sc.Config.CSRFOptions.FieldName,
			Token:      csrf.Token(c.Request()),
		})
	}
}

type LoginData struct {
	NoAuth        bool
	ReturnURL     string
	ErrorMessages []string
}

func (s *Service) getLogin() echo.HandlerFunc {
	t := "auth/login.page.tmpl"
	s.r.MustHave(t)
	return func(c echo.Context) error {
		// Check authentication & authorization
		a, sess, err := auth.GetWithSession(c, s.sc)
		if err != nil {
			return err
		}

		// Consume & save session
		errorMessages, err := session.GetErrorMessages(sess)
		if err != nil {
			return err
		}
		loginData := LoginData{
			NoAuth:        s.ac.Config.NoAuth,
			ReturnURL:     c.QueryParam("return"),
			ErrorMessages: errorMessages,
		}
		if err = session.Save(sess, c); err != nil {
			return err
		}

		// Add non-persistent overrides of session data
		a.CSRF = auth.OverrideCSRFInlining(c.Request(), a.CSRF, true)

		// Produce output
		return s.r.CacheablePage(c.Response(), c.Request(), t, loginData, a)
	}
}

func sanitizeReturnURL(returnURL string) (*url.URL, error) {
	u, err := url.ParseRequestURI(returnURL)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func handleAuthenticationSuccess(
	c echo.Context, username, returnURL string, omitCSRFToken bool, sc *sessions.Client,
) error {
	// Update session
	sess, err := sc.Regenerate(c)
	if err != nil {
		return err
	}
	auth.SetIdentity(sess, username)
	// This allows client-side Javascript to specify for server-side session data that we only need
	// to provide CSRF tokens through the /csrf route and we can omit them from HTML response
	// bodies, in order to make HTML responses cacheable.
	auth.SetCSRFBehavior(sess, !omitCSRFToken)
	if err = session.Save(sess, c); err != nil {
		return err
	}

	// Redirect user
	u, err := sanitizeReturnURL(returnURL)
	if err != nil {
		// TODO: log the error, too
		return c.Redirect(http.StatusSeeOther, "/")
	}
	return c.Redirect(http.StatusSeeOther, u.String())
}

func handleAuthenticationFailure(c echo.Context, returnURL string, sc *sessions.Client) error {
	// Update session
	sess, serr := sc.Get(c)
	if serr != nil {
		return serr
	}
	session.AddErrorMessage(sess, "Could not log in!")
	auth.SetIdentity(sess, "")
	if err := session.Save(sess, c); err != nil {
		return err
	}

	// Redirect user
	u, err := sanitizeReturnURL(returnURL)
	if err != nil {
		// TODO: log the error, too
		return c.Redirect(http.StatusSeeOther, "/login")
	}
	r := url.URL{Path: "/login"}
	q := r.Query()
	q.Set("return", u.String())
	r.RawQuery = q.Encode()
	return c.Redirect(http.StatusSeeOther, r.String())
}

func (s *Service) postSessions() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Parse params
		state := c.FormValue("state")

		// Run queries
		switch state {
		default:
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"invalid session %s", state,
			))
		case "authenticated":
			username := c.FormValue("username")
			password := c.FormValue("password")
			returnURL := c.FormValue("return")
			omitCSRFToken := strings.ToLower(c.FormValue("omit-csrf-token")) == "true"

			// TODO: add session attacks detection. Refer to the "Session Attacks Detection" section of
			// the OWASP Session Management Cheat Sheet

			// Check authentication
			identified, err := s.ac.CheckCredentials(username, password)
			if err != nil {
				return err
			}
			if !identified {
				return handleAuthenticationFailure(c, returnURL, s.sc)
			}
			return handleAuthenticationSuccess(c, username, returnURL, omitCSRFToken, s.sc)
		case "unauthenticated":
			// TODO: add a client-side controller to automatically submit a logout request after the
			// idle timeout expires, and display an inactivity logout message
			sess, err := s.sc.Invalidate(c)
			if err != nil {
				return err
			}
			if err := session.Save(sess, c); err != nil {
				return err
			}
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, "/")
	}
}

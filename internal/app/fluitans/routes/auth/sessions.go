package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/session"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
)

type CSRFViewData struct {
	HeaderName string `json:"headerName,omitempty"`
	FieldName  string `json:"fieldName,omitempty"`
	Token      string `json:"token,omitempty"`
}

func (h *Handlers) HandleCSRFGet() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Produce output
		godest.WithUncacheable()(c.Response().Header())
		return c.JSON(http.StatusOK, CSRFViewData{
			HeaderName: h.ss.CSRFOptions().HeaderName,
			FieldName:  h.ss.CSRFOptions().FieldName,
			Token:      csrf.Token(c.Request()),
		})
	}
}

type LoginViewData struct {
	NoAuth        bool
	ReturnURL     string
	ErrorMessages []string
}

func (h *Handlers) HandleLoginGet() auth.HTTPHandlerFuncWithSession {
	t := "auth/login.page.tmpl"
	h.r.MustHave(t)
	return func(c echo.Context, a auth.Auth, sess *sessions.Session) error {
		// Consume & save session
		errorMessages, err := session.GetErrorMessages(sess)
		if err != nil {
			return err
		}
		loginViewData := LoginViewData{
			NoAuth:        h.ac.Config.NoAuth,
			ReturnURL:     c.QueryParam("return"),
			ErrorMessages: errorMessages,
		}
		if err := sess.Save(c.Request(), c.Response()); err != nil {
			return err
		}

		// Add non-persistent overrides of session data
		a.CSRF.SetInlining(c.Request(), true)

		// Produce output
		return h.r.CacheablePage(c.Response(), c.Request(), t, loginViewData, a)
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
	c echo.Context, username, returnURL string, omitCSRFToken bool, ss *session.Store,
) error {
	// Update session
	sess, err := ss.Get(c.Request())
	if err != nil {
		return err
	}
	session.Regenerate(sess)
	auth.SetIdentity(sess, username)
	// This allows client-side Javascript to specify for server-side session data that we only need
	// to provide CSRF tokens through the /csrf route and we can omit them from HTML response
	// bodies, in order to make HTML responses cacheable.
	auth.SetCSRFBehavior(sess, !omitCSRFToken)
	if err = sess.Save(c.Request(), c.Response()); err != nil {
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

func handleAuthenticationFailure(c echo.Context, returnURL string, ss *session.Store) error {
	// Update session
	sess, serr := ss.Get(c.Request())
	if serr != nil {
		return serr
	}
	session.AddErrorMessage(sess, "Could not log in!")
	auth.SetIdentity(sess, "")
	if err := sess.Save(c.Request(), c.Response()); err != nil {
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

func (h *Handlers) HandleSessionsPost() echo.HandlerFunc {
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
			identified, err := h.ac.CheckCredentials(username, password)
			if err != nil {
				return err
			}
			if !identified {
				return handleAuthenticationFailure(c, returnURL, h.ss)
			}
			return handleAuthenticationSuccess(c, username, returnURL, omitCSRFToken, h.ss)
		case "unauthenticated":
			// TODO: add a client-side controller to automatically submit a logout request after the
			// idle timeout expires, and display an inactivity logout message
			sess, err := h.ss.Get(c.Request())
			if err != nil {
				return err
			}
			h.acc.Cancel(sess.ID)
			session.Invalidate(sess)
			if err := sess.Save(c.Request(), c.Response()); err != nil {
				return err
			}
		}

		// Redirect user
		return c.Redirect(http.StatusSeeOther, "/")
	}
}

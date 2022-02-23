package auth

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/framework/session"
)

type CSRFData struct {
	HeaderName string `json:"headerName,omitempty"`
	FieldName  string `json:"fieldName,omitempty"`
	Token      string `json:"token,omitempty"`
}

func getCSRF(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	app, ok := g.App.(*client.Globals)
	if !ok {
		return nil, client.NewUnexpectedGlobalsTypeError(g.App)
	}
	sc := app.Clients.Sessions
	return func(c echo.Context) error {
		// Get session
		sess, err := sc.Get(c)
		if err != nil {
			return err
		}
		if err = session.Save(sess, c); err != nil {
			return err
		}

		// Produce output
		return c.JSON(http.StatusOK, CSRFData{
			HeaderName: sc.Config.CSRFOptions.HeaderName,
			FieldName:  sc.Config.CSRFOptions.FieldName,
			Token:      csrf.Token(c.Request()),
		})
	}, nil
}

type LoginData struct {
	NoAuth        bool
	ReturnURL     string
	ErrorMessages []string
}

func getLogin(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "auth/login.page.tmpl"
	err := te.RequireSegments("auth.getLogin", t)
	if err != nil {
		return nil, err
	}

	app, ok := g.App.(*client.Globals)
	if !ok {
		return nil, client.NewUnexpectedGlobalsTypeError(g.App)
	}
	return func(c echo.Context) error {
		// Check authentication & authorization
		a, sess, err := auth.GetWithSession(c, app.Clients.Sessions)
		if err != nil {
			return err
		}

		// Consume & save session
		errorMessages, err := session.GetErrorMessages(sess)
		if err != nil {
			return err
		}
		loginData := LoginData{
			NoAuth:        app.Clients.Authn.Config.NoAuth,
			ReturnURL:     c.QueryParam("return"),
			ErrorMessages: errorMessages,
		}
		if err = session.Save(sess, c); err != nil {
			return err
		}

		// Produce output
		return route.Render(c, t, loginData, a, te, g)
	}, nil
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
	auth.SetCSRFBehavior(sess, omitCSRFToken)
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

func postSessions(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	app, ok := g.App.(*client.Globals)
	if !ok {
		return nil, client.NewUnexpectedGlobalsTypeError(g.App)
	}
	sc := app.Clients.Sessions
	return func(c echo.Context) error {
		// Parse params
		formAction := c.FormValue("form-action")

		// Run queries
		switch formAction {
		default:
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"invalid POST form action %s", formAction,
			))
		case "create":
			username := c.FormValue("username")
			password := c.FormValue("password")
			returnURL := c.FormValue("return")
			omitCSRFToken := strings.ToLower(c.FormValue("omit-csrf-token")) == "true"

			// TODO: add session attacks detection. Refer to the "Session Attacks Detection" section of
			// the OWASP Session Management Cheat Sheet

			// Check authentication
			identified, err := app.Clients.Authn.CheckCredentials(username, password)
			if err != nil {
				return err
			}
			if !identified {
				return handleAuthenticationFailure(c, returnURL, sc)
			}
			return handleAuthenticationSuccess(c, username, returnURL, omitCSRFToken, sc)
		case "delete":
			// TODO: add a client-side controller to automatically submit a logout request after the
			// idle timeout expires, and display an inactivity logout message
			sess, err := sc.Invalidate(c)
			if err != nil {
				return err
			}
			if err := session.Save(sess, c); err != nil {
				return err
			}
		}

		return c.Redirect(http.StatusSeeOther, "/")
	}, nil
}

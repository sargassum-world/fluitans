package auth

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/framework/session"
)

type LoginData struct {
	NoAuth        bool
	ReturnURL     string
	ErrorMessages []string
}

func getLogin(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "auth/login.page.tmpl"
	err := te.RequireSegments("authn.getLogin", t)
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
	c echo.Context, username, returnURL string, sc *sessions.Client,
) error {
	// Update session
	sess, err := sc.Regenerate(c)
	if err != nil {
		return err
	}
	auth.SetIdentity(sess, username)
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
		method := c.FormValue("method")

		// Run queries
		switch method {
		default:
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"invalid POST method %s", method,
			))
		case "AUTHENTICATE":
			username := c.FormValue("username")
			password := c.FormValue("password")
			returnURL := c.FormValue("return")
			// TODO: add intrusion detection
			identified, err := app.Clients.Authn.CheckCredentials(username, password)
			if err != nil {
				return err
			}
			if !identified {
				return handleAuthenticationFailure(c, returnURL, sc)
			}
			return handleAuthenticationSuccess(c, username, returnURL, sc)
		case "DELETE":
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

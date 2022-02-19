// Package routes contains the route handlers for the Fluitans server.
package routes

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/framework/session"
)

var AuthnPages = []route.Templated{
	{
		Path:         "/login",
		Method:       http.MethodGet,
		HandlerMaker: getLogin,
		Templates:    []string{"auth/login.page.tmpl"},
	},
	{
		Path:         "/sessions",
		Method:       http.MethodPost,
		HandlerMaker: postSessions,
		Templates:    []string{},
	},
}

type LoginData struct {
	NoAuth        bool
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
			ErrorMessages: errorMessages,
		}
		if err = session.Save(sess, c); err != nil {
			return err
		}

		// Produce output
		return route.Render(c, t, loginData, a, te, g)
	}, nil
}

func handleAuthenticationSuccess(ctx echo.Context, username string, sc *sessions.Client) error {
	sess, err := sc.Regenerate(ctx)
	if err != nil {
		return err
	}

	// TODO: add intrusion detection
	auth.SetIdentity(sess, username)
	if err = session.Save(sess, ctx); err != nil {
		return err
	}
	// TODO: redirect to the previous page by getting the path from a GET parameter
	return nil
}

func handleAuthenticationFailure(ctx echo.Context, sc *sessions.Client) error {
	sess, serr := sc.Get(ctx)
	if serr != nil {
		return serr
	}

	session.AddErrorMessage(sess, "Could not log in!")
	auth.SetIdentity(sess, "")
	if err := session.Save(sess, ctx); err != nil {
		return err
	}
	// TODO: preserve GET redirect parameter
	return ctx.Redirect(http.StatusSeeOther, "/login")
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
			identified, err := app.Clients.Authn.CheckCredentials(username, password)
			if err != nil {
				return err
			}
			if identified {
				return handleAuthenticationSuccess(c, username, sc)
			}
			return handleAuthenticationFailure(c, sc)
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

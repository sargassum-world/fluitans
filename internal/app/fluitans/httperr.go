// Package fluitans provides the Fluitans server.
package fluitans

import (
	"fmt"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/pkg/framework"
	"github.com/sargassum-eco/fluitans/pkg/framework/session"
)

type ErrorData struct {
	Code     int
	Error    framework.HTTPError
	Messages []string
}

func NewHTTPErrorHandler(tr framework.TemplateRenderer, sc *sessions.Client) echo.HTTPErrorHandler {
	tr.MustHave("app/httperr.page.tmpl")
	return func(err error, c echo.Context) {
		c.Logger().Error(err)

		// Check authentication & authorization
		a, sess, serr := auth.GetWithSession(c, sc)
		if serr != nil {
			c.Logger().Error(errors.Wrap(serr, "couldn't get session+auth in error handler"))
		}

		// Process error code
		code := http.StatusInternalServerError
		if herr, ok := err.(*echo.HTTPError); ok {
			code = herr.Code
		}
		errorData := ErrorData{
			Code:  code,
			Error: framework.DescribeHTTPError(code),
		}

		// Consume & save session
		if sess != nil {
			messages, merr := session.GetErrorMessages(sess)
			if merr != nil {
				c.Logger().Error(errors.Wrap(
					merr, "couldn't get error messages from session in error handler",
				))
			}
			errorData.Messages = messages
			if serr = session.Save(sess, c); serr != nil {
				c.Logger().Error(errors.Wrap(serr, "couldn't save session in error handler"))
			}
		}

		// Produce output
		tr.SetUncacheable(c.Response().Header())
		if perr := tr.Page(
			c.Response(), c.Request(), code, "app/httperr.page.tmpl", errorData, a,
		); perr != nil {
			c.Logger().Error(errors.Wrap(perr, "couldn't render error page in error handler"))
		}
	}
}

func NewCSRFErrorHandler(
	tr framework.TemplateRenderer, l echo.Logger, sc *sessions.Client,
) http.HandlerFunc {
	tr.MustHave("app/httperr.page.tmpl")
	return func(w http.ResponseWriter, r *http.Request) {
		l.Error(csrf.FailureReason(r))
		// Check authentication & authorization
		sess, serr := session.Get(r, sc.Config.CookieName, sc.Store)
		if serr != nil {
			l.Error(errors.Wrap(serr, "couldn't get session in error handler"))
		}
		var a auth.Auth
		if sess != nil {
			a, serr = auth.GetFromRequest(r, *sess, sc)
			if serr != nil {
				l.Error(errors.Wrap(serr, "couldn't get auth in error handler"))
			}
		}

		// Generate error code
		code := http.StatusForbidden
		errorData := ErrorData{
			Code:  code,
			Error: framework.DescribeHTTPError(code),
			Messages: []string{
				fmt.Sprintf(
					"%s. If you disabled Javascript after signing in, "+
						"please clear your cookies for this site and sign in again.",
					csrf.FailureReason(r).Error(),
				),
			},
		}

		// Produce output
		tr.SetUncacheable(w.Header())
		if rerr := tr.Page(w, r, code, "app/httperr.page.tmpl", errorData, a); rerr != nil {
			l.Error(errors.Wrap(rerr, "couldn't render error page in error handler"))
		}
	}
}

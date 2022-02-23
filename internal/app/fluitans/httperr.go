// Package fluitans provides the Fluitans server.
package fluitans

import (
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/pkg/framework/httperr"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/framework/session"
	"github.com/sargassum-eco/fluitans/pkg/framework/template"
)

type ErrorData struct {
	Code     int
	Error    httperr.Error
	Messages []string
}

func NewHTTPErrorHandler(
	tg route.TemplateGlobals, sc *sessions.Client,
) echo.HTTPErrorHandler {
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
			Error: httperr.DescribeError(code),
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

		// Render error page
		if perr := c.Render(
			code, "app/httperr.page.tmpl", route.NewRenderData(c.Request(), tg, errorData, a),
		); perr != nil {
			c.Logger().Error(errors.Wrap(perr, "couldn't render error page in error handler"))
		}
	}
}

func NewCSRFErrorHandler(
	tg route.TemplateGlobals, renderer *template.TemplateRenderer, l echo.Logger, sc *sessions.Client,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l.Error(csrf.FailureReason(r))
		// Check authentication & authorization
		sess, serr := session.Get(r, sc.Config.CookieName, sc.Store)
		if serr != nil {
			l.Error(errors.Wrap(serr, "couldn't get session in error handler"))
		}

		var a auth.Auth
		if sess != nil {
			a, serr = auth.GetFromRequest(r, *sess)
			if serr != nil {
				l.Error(errors.Wrap(serr, "couldn't get auth in error handler"))
			}
		}

		// Generate error code
		code := http.StatusForbidden
		errorData := ErrorData{
			Code:     code,
			Error:    httperr.DescribeError(code),
			Messages: []string{csrf.FailureReason(r).Error()},
		}

		if rerr := route.WriteTemplatedResponse(
			w, r, renderer, "app/httperr.page.tmpl", code, errorData, a, tg,
		); rerr != nil {
			l.Error(errors.Wrap(rerr, "couldn't render error page in error handler"))
		}
	}
}

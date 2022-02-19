// Package fluitans provides the Fluitans server.
package fluitans

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/clients/sessions"
	"github.com/sargassum-eco/fluitans/pkg/framework/httperr"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
	"github.com/sargassum-eco/fluitans/pkg/framework/session"
)

type ErrorData struct {
	Code int
	Error httperr.Error
	Messages []string
}

func NewHTTPErrorHandler(
	tg route.TemplateGlobals, sc *sessions.Client,
) (func(err error, c echo.Context), error) {
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
			Code: code,
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
			if serr = sess.Save(c.Request(), c.Response()); serr != nil {
				c.Logger().Error(errors.Wrap(serr, "couldn't save session in error handler"))
			}
		}

		// Render error page
		perr := c.Render(
			code, "app/httperr.page.tmpl", route.MakeRenderData(c, tg, errorData, a),
		)
		if perr != nil {
			c.Logger().Error(errors.Wrap(perr, "couldn't render error page in error handler"))
		}
	}, nil
}

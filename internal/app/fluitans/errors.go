package fluitans

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/template"
)

func NewHTTPErrorHandler() (func(err error, c echo.Context), error) {
	tg, _, err := computeGlobals()
	if err != nil {
		return nil, err
	}

	g := *tg
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if herr, ok := err.(*echo.HTTPError); ok {
			code = herr.Code
		}
		perr := c.Render(code, "httperr.page.tmpl", struct {
			Meta      template.Meta
			Embeds    template.Embeds
			ErrorCode int
		}{
			Meta: template.Meta{
				Path: c.Request().URL.Path,
			},
			Embeds:    g.Embeds,
			ErrorCode: code,
		})
		if perr != nil {
			c.Logger().Error(err)
		}
		c.Logger().Error(err)
	}, nil
}

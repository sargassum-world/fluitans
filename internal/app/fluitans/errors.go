package fluitans

import (
	"net/http"
	"os"

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
			Meta   template.Meta
			Embeds template.Embeds
			Data   int
		}{
			Meta: template.Meta{
				Path:       c.Request().URL.Path,
				DomainName: os.Getenv("FLUITANS_DOMAIN_NAME"),
			},
			Embeds: g.Embeds,
			Data:   code,
		})
		if perr != nil {
			c.Logger().Error(err)
		}
		c.Logger().Error(err)
	}, nil
}

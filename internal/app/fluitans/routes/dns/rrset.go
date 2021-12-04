package dns

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

func postRRset(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	switch app := g.App.(type) {
	default:
		return nil, client.NewUnexpectedGlobalsTypeError(app)
	case *client.Globals:
		return func(c echo.Context) error {
			// Extract context
			ctx := c.Request().Context()

			// Parse params
			subname := c.Param("subname")
			if subname == "@" {
				subname = ""
			}
			recordType := c.Param("type")
			method := c.FormValue("method")

			// Run queries
			switch method {
			default:
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
					"invalid POST method %s", method,
				))
			case "DELETE":
				if err := app.Clients.Desec.DeleteRRset(ctx, subname, recordType); err != nil {
					return err
				}

				return c.Redirect(http.StatusSeeOther, "/dns")
			}
		}, nil
	}
}

package dns

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

func postRRset(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	app, ok := g.App.(*client.Globals)
	if !ok {
		return nil, client.NewUnexpectedGlobalsTypeError(g.App)
	}
	return func(c echo.Context) error {
		// Check authentication & authorization
		if err := auth.RequireAuthorized(c, app.Clients.Sessions); err != nil {
			return err
		}

		// Extract context
		ctx := c.Request().Context()

		// Parse params
		subname := c.Param("subname")
		if subname == "@" {
			subname = ""
		}
		recordType := c.Param("type")
		formAction := c.FormValue("form-action")

		// Run queries
		switch formAction {
		default:
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf(
				"invalid POST form action %s", formAction,
			))
		case "delete":
			if err := app.Clients.Desec.DeleteRRset(ctx, subname, recordType); err != nil {
				return err
			}

			return c.Redirect(http.StatusSeeOther, "/dns")
		}
	}, nil
}

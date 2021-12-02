package dns

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/route"
)

func postRRset(
	g route.TemplateGlobals, te route.TemplateEtagSegments,
) (echo.HandlerFunc, error) {
	switch app := g.App.(type) {
	default:
		return nil, errors.Errorf("app globals are of unexpected type %T", g.App)
	case *client.Globals:
		return func(c echo.Context) error {
			// Parse params
			subname := c.Param("subname")
			if subname == "@" {
				subname = ""
			}
			recordType := c.Param("type")
			method := c.FormValue("method")

			// Run queries
			domain, err := client.NewDNSDomain(
				app.RateLimiters[client.DesecReadLimiterName], app.Cache,
			)
			if err != nil {
				return err
			}

			switch method {
			case "DELETE":
				if err = client.DeleteRRset(c, *domain, subname, recordType); err != nil {
					return err
				}

				return c.Redirect(http.StatusSeeOther, "/dns")
			}

			return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/dns/%s...", subname))
		}, nil
	}
}

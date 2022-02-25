package dns

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-eco/fluitans/pkg/framework/route"
)

func (s *Service) postRRset(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	return func(c echo.Context) error {
		// Check authentication & authorization
		if err := auth.RequireAuthorized(c, s.sc); err != nil {
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
		state := c.FormValue("state")

		// Run queries
		switch state {
		default:
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("invalid RRset state %s", state))
		case "deleted":
			if err := s.dc.DeleteRRset(ctx, subname, recordType); err != nil {
				return err
			}

			return c.Redirect(http.StatusSeeOther, "/dns")
		}
	}, nil
}

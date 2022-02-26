package dns

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/auth"
)

func (s *Service) postRRset() echo.HandlerFunc {
	return func(c echo.Context) error {
		// Check authentication & authorization
		if err := auth.RequireAuthorized(c, s.sc); err != nil {
			return err
		}

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
			if err := s.dc.DeleteRRset(c.Request().Context(), subname, recordType); err != nil {
				return err
			}

			// Redirect user
			return c.Redirect(http.StatusSeeOther, "/dns")
		}
	}
}

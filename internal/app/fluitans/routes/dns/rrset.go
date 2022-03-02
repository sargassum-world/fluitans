package dns

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (h *Handlers) HandleRRsetPost() echo.HandlerFunc {
	return func(c echo.Context) error {
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
			if err := h.dc.DeleteRRset(c.Request().Context(), subname, recordType); err != nil {
				return err
			}

			// Redirect user
			return c.Redirect(http.StatusSeeOther, "/dns")
		}
	}
}

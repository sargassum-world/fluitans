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

			// We don't return a Turbo Stream, because an RRset deletion changes the contents of the
			// dns/domain.partial.tmpl and subdomain.partial.tmpl templates beyond the RRset partial
			// itself, and then we'd have to render different Turbo Streams for each
			// possible parent of the RRset partial. For now, it's not worth the complexity.
			// Redirect user
			return c.Redirect(http.StatusSeeOther, "/dns")
		}
	}
}

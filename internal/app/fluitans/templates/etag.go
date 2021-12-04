package templates

import (
	"encoding/json"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/caching"
	"github.com/sargassum-eco/fluitans/internal/fingerprint"
)

func ProcessEtag(
	c echo.Context, templateEtagSegments []string, data interface{},
) (bool, error) {
	marshaled, err := json.Marshal(data)
	if err != nil {
		return false, err
	}

	return caching.ProcessEtag(
		c, append(templateEtagSegments, fingerprint.Compute(marshaled)),
	)
}

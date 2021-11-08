package networks

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/caching"
	"github.com/sargassum-eco/fluitans/internal/fingerprint"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
)

func getControllers(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "networks/controllers.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	return func(c echo.Context) error {
		// Run queries
		controllers, err := client.GetControllers()
		if err != nil {
			return err
		}

		// Handle Etag
		data, err := json.Marshal(controllers)
		if err != nil {
			return err
		}

		if noContent, err := caching.ProcessEtag(c, tte, fingerprint.Compute(data)); noContent {
			return err
		}

		// Render template
		return c.Render(http.StatusOK, t, struct {
			Meta   template.Meta
			Embeds template.Embeds
			Data   []client.Controller
		}{
			Meta: template.Meta{
				Path:       c.Request().URL.Path,
				DomainName: client.GetDomainName(),
			},
			Embeds: g.Embeds,
			Data:   controllers,
		})
	}, nil
}

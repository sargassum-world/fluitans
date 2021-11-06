package networks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"

	"github.com/sargassum-eco/fluitans/internal/caching"
	"github.com/sargassum-eco/fluitans/internal/fingerprint"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
	"github.com/sargassum-eco/fluitans/pkg/zerotier"
)

type Controller struct {
	// TODO: move this to a models package or something!
	Server      string `json:"server"`
	Name        string `json:"name"` // Must be unique for display purposes!
	Description string `json:"description"`
	Authtoken   string `json:"authtoken"`
}

// TODO: move this into globals!
// TODO: store these records in a secrets file instead of hard-coding them!
// TODO: store these records in the database instead of on the filesystem!
var storedControllers = []Controller{
	{
		Server:      os.Getenv("FLUITANS_ZT_CONTROLLER_SERVER"),
		Name:        "zerotier-test",
		Description: "An insecure ZeroTier network controller used purely for development and testing.",
		Authtoken:   os.Getenv("FLUITANS_ZT_CONTROLLER_AUTHTOKEN"),
	},
}

func findController(controllers []Controller, name string) (*Controller, bool) {
	found := false
	for _, v := range controllers {
		if v.Name == name {
			return &v, true
		}
	}
	return nil, found
}

func findControllerByAddress(
	c echo.Context, controllers []Controller, address string,
) (*Controller, error) {
	// TODO: we should instead first look up the address in a cache and then
	// issue a request to the controller to verify it still has the address;
	// if not, we should update the cache. If the address isn't in the cache,
	// then we should query all controllers, starting with the ones not in
	// the cache
	eg, ctx := errgroup.WithContext(c.Request().Context())
	addresses := make([]string, len(controllers))
	for i, controller := range controllers {
		eg.Go(func(i int) func() error {
			return func() error {
				client, cerr := zerotier.NewAuthClientWithResponses(
					controller.Server, controller.Authtoken,
				)
				if cerr != nil {
					return nil
				}

				res, err := client.GetStatusWithResponse(ctx)
				if err != nil {
					return err
				}

				addresses[i] = *res.JSON200.Address
				return nil
			}
		}(i))
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	for i, v := range controllers {
		if addresses[i] == address {
			return &v, nil
		}
	}

	return nil, echo.NewHTTPError(
		http.StatusNotFound,
		fmt.Sprintf("Controller not found with address %s", address),
	)
}

func controllers(g route.TemplateGlobals, te route.TemplateEtagSegments) (echo.HandlerFunc, error) {
	t := "networks/controllers.page.tmpl"
	tte, ok := te[t]
	if !ok {
		return nil, te.NewNotFoundError(t)
	}

	data, err := json.Marshal(storedControllers)
	if err != nil {
		return nil, err
	}

	return func(c echo.Context) error {
		// Handle Etag
		if noContent, err := caching.ProcessEtag(
			c,
			tte,
			fingerprint.Compute(data),
		); noContent {
			return err
		}

		// Render template
		return c.Render(http.StatusOK, t, struct {
			Meta   template.Meta
			Embeds template.Embeds
			Data   []Controller
		}{
			Meta: template.Meta{
				Path:       c.Request().URL.Path,
				DomainName: os.Getenv("FLUITANS_DOMAIN_NAME"),
			},
			Embeds: g.Embeds,
			Data:   storedControllers,
		})
	}, nil
}

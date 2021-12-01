// Package fluitans provides the Fluitans server.
package fluitans

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/routes"
	"github.com/sargassum-eco/fluitans/internal/app/fluitans/templates"
	"github.com/sargassum-eco/fluitans/internal/route"
	"github.com/sargassum-eco/fluitans/internal/template"
	"github.com/sargassum-eco/fluitans/web"
)

func NewRenderer() *templates.TemplateRenderer {
	return templates.New(web.AppHFS.HashName, web.StaticHFS.HashName, web.TemplatesFS)
}

func launchBackgroundWorkers(ag *client.Globals) {
	go func() {
		var writeInterval time.Duration = 5000
		writeLimiter := ag.RateLimiters[client.DesecWriteLimiterName]
		for {
			if writeLimiter.TryAdd(time.Now(), 1) {
				/*fmt.Printf(
					"Bumped the write limiter: %+v\n",
					writeLimiter.EstimateFillRatios(time.Now()),
				)*/
			} else {
				fmt.Printf(
					"Write limiter throttled: wait %f sec\n",
					writeLimiter.EstimateWaitDuration(time.Now(), 1).Seconds(),
				)
			}
			time.Sleep(writeInterval * time.Millisecond)
		}
	}()
}

func RegisterRoutes(e route.EchoRouter) (*Globals, error) {
	f, err := computeTemplateFingerprints()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't compute template fingerprints")
	}

	ag, err := makeAppGlobals()
	if err != nil {
		return nil, errors.Wrap(err, "couldn't make app globals")
	}

	launchBackgroundWorkers(ag)

	g := &Globals{
		Template: route.TemplateGlobals{
			Embeds:               makeTemplateEmbeds(),
			TemplateFingerprints: *f,
			App:                  ag,
		},
		Static: makeStaticGlobals(),
	}

	err = route.RegisterTemplated(e, routes.TemplatedAssets, g.Template)
	if err != nil {
		return g, errors.Wrap(err, "couldn't register templated assets")
	}

	err = route.RegisterStatic(e, routes.StaticAssets, g.Static)
	if err != nil {
		return g, errors.Wrap(err, "couldn't register static assets")
	}

	err = route.RegisterTemplated(e, routes.Pages, g.Template)
	if err != nil {
		return g, errors.Wrap(err, "couldn't register templated routes")
	}

	return g, nil
}

func NewHTTPErrorHandler(g *Globals) func(err error, c echo.Context) {
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		if herr, ok := err.(*echo.HTTPError); ok {
			code = herr.Code
		}
		perr := c.Render(code, "httperr.page.tmpl", struct {
			Meta   template.Meta
			Embeds template.Embeds
			Data   int
		}{
			Meta: template.Meta{
				Path:       c.Request().URL.Path,
				DomainName: client.GetEnvVarDomainName(),
			},
			Embeds: g.Template.Embeds,
			Data:   code,
		})
		if perr != nil {
			c.Logger().Error(err)
		}
		c.Logger().Error(err)
	}
}

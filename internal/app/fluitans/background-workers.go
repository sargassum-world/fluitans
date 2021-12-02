package fluitans

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"

	"github.com/sargassum-eco/fluitans/internal/app/fluitans/client"
)

func prescanZerotierControllers(e *echo.Echo, cg *client.Globals) {
	// TODO: instead of making a new echo.Context object, instead change the client
	// functions to not expect an echo.Context object
	request, err := http.NewRequestWithContext(
		context.Background(), "GET", "", strings.NewReader(""),
	)
	if err != nil {
		e.Logger.Error(errors.Wrap(err, "couldn't make dummy request to make dummy echo.Context"))
	}

	responseWriter := httptest.NewRecorder()
	c := e.NewContext(request, responseWriter)
	controllers, err := client.GetControllers()
	if err != nil {
		e.Logger.Error(errors.Wrap(err, "couldn't get the list of known controllers"))
		return
	}

	_, err = client.ScanControllers(c, controllers, cg.Cache)
	if err != nil {
		e.Logger.Error(errors.Wrap(err, "couldn't prescan Zerotier controllers for cache"))
	}
}

func prefetchDNSRecords(e *echo.Echo, cg *client.Globals) {
	// TODO: instead of making a new echo.Context object, instead change the client
	// functions to not expect an echo.Context object
	request, err := http.NewRequestWithContext(
		context.Background(), "GET", "", strings.NewReader(""),
	)
	if err != nil {
		e.Logger.Error(errors.Wrap(err, "couldn't make dummy request to make dummy echo.Context"))
	}

	responseWriter := httptest.NewRecorder()
	c := e.NewContext(request, responseWriter)
	domain, err := client.NewDNSDomain(
		cg.RateLimiters[client.DesecReadLimiterName], cg.Cache,
	)
	if err != nil {
		e.Logger.Error(errors.Wrap(err, "couldn't make DNS Domain client object"))
		return
	}

	_, err = client.GetRRsets(c, *domain)
	if err != nil {
		e.Logger.Error(errors.Wrap(err, "couldn't prefetch DNS records for cache"))
	}
}

func testWriteLimiter(cg *client.Globals) {
	var writeInterval time.Duration = 5000
	writeLimiter := cg.RateLimiters[client.DesecWriteLimiterName]
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
}

func LaunchBackgroundWorkers(e *echo.Echo, g *Globals) error {
	switch app := g.Template.App.(type) {
	default:
		return fmt.Errorf("app globals are of unexpected type %T", g.Template.App)
	case *client.Globals:
		go prescanZerotierControllers(e, app)
		go prefetchDNSRecords(e, app)
		go testWriteLimiter(app)
	}
	return nil
}

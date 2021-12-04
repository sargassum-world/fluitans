package route

import (
	"fmt"
	"io/fs"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

type StaticGlobals struct {
	FS  map[string]fs.FS
	HFS map[string]fs.FS
}

type Static struct {
	Path         string
	HandlerMaker func(g StaticGlobals) (echo.HandlerFunc, error)
}

func RegisterStatic(e EchoRouter, routes []Static, g StaticGlobals) error {
	for _, route := range routes {
		h, err := route.HandlerMaker(g)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf(
				"couldn't make the handler for static route %s", route.Path),
			)
		}

		e.GET(route.Path, h)
	}
	return nil
}

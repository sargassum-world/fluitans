package route

import (
	"io/fs"

	"github.com/labstack/echo/v4"
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
			return err
		}

		e.GET(route.Path, h)
	}
	return nil
}

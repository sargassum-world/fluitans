package godest

import (
	"io"

	"github.com/labstack/echo/v4"

	"github.com/sargassum-eco/fluitans/pkg/godest/template"
)

// Routing

// EchoRouter is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration.
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// Templates

type EchoRenderer struct {
	t template.Templates
}

func NewEchoRenderer(t template.Templates) EchoRenderer {
	return EchoRenderer{
		t: t,
	}
}

func (r EchoRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return r.t.Execute(w, name, data)
}

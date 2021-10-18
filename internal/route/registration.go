// Package route enables declarative routing for Echo.
package route

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

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
type RegistrationFunc func(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route

func GetRegistrationFuncs(e EchoRouter) map[string]RegistrationFunc {
	return map[string]RegistrationFunc{
		http.MethodGet:     e.GET,
		http.MethodHead:    e.HEAD,
		http.MethodPost:    e.POST,
		http.MethodPut:     e.PUT,
		http.MethodPatch:   e.PATCH,
		http.MethodDelete:  e.DELETE,
		http.MethodConnect: e.CONNECT,
		http.MethodOptions: e.OPTIONS,
		http.MethodTrace:   e.TRACE,
	}
}

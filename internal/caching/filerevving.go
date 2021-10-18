// Package caching provides utilities for working with HTTP caching
// in Echo
package caching

import (
	"fmt"

	"github.com/labstack/echo/v4"
)

func WrapStaticHeader(h echo.HandlerFunc, age int) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", age))
		return h(c)
	}
}

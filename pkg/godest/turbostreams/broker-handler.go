package turbostreams

import (
	"reflect"
	"runtime"

	"github.com/pkg/errors"
)

// Methods

const (
	MethodPub   = "PUB"
	MethodSub   = "SUB"
	MethodUnsub = "UNSUB"
	MethodMsg   = "MSG"
)

// Handlers

type HandlerFunc func(c Context) error

func NotFoundHandler(c Context) error {
	return errors.Errorf("handler not found for topic %s", c.Topic())
}

func EmptyHandler(c Context) error {
	return nil
}

type methodHandler struct {
	pub   HandlerFunc
	sub   HandlerFunc
	unsub HandlerFunc
	msg   HandlerFunc
}

func (m *methodHandler) isHandler() bool {
	return m.pub != nil || m.sub != nil || m.unsub != nil || m.msg != nil
}

func handlerName(h HandlerFunc) string {
	// Copied from github.com/labstack/echo's handlerName function
	t := reflect.ValueOf(h).Type()
	if t.Kind() == reflect.Func {
		return runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	}
	return t.String()
}

// Middleware

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

func applyMiddleware(h HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

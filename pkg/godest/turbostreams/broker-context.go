package turbostreams

import (
	stdContext "context"

	"github.com/sargassum-world/fluitans/pkg/godest/pubsub"
)

type Context interface {
	Context() stdContext.Context
	Topic() string
	SessionID() string
	Param(name string) string
	Broadcast(topic, message string)
	// TODO: rename to Data, return an interface{}
	Message() string
	// TODO: replace this with a Message method
	Write(message string)
}

type context struct {
	context   stdContext.Context
	path      string
	pnames    []string
	pvalues   []string
	handler   HandlerFunc
	hub       *pubsub.StringHub
	topic     string
	sessionID string
	message   string
}

func (c *context) Context() stdContext.Context {
	return c.context
}

func (c *context) Topic() string {
	return c.topic
}

func (c *context) SessionID() string {
	return c.sessionID
}

func (c *context) Param(name string) string {
	// Copied from github.com/labstack/echo's context.Param method
	for i, n := range c.pnames {
		if i < len(c.pvalues) {
			if n == name {
				return c.pvalues[i]
			}
		}
	}
	return ""
}

func (c *context) Broadcast(topic, message string) {
	c.hub.Broadcast(topic, message)
}

func (c *context) Message() string {
	return c.message
}

func (c *context) Write(message string) {
	c.message = message
}

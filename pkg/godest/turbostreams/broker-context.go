package turbostreams

import (
	"bytes"
	stdContext "context"
	"io"
)

type Context interface {
	Context() stdContext.Context
	Topic() string
	SessionID() string
	Param(name string) string
	Hub() *MessagesHub
	Publish(messages ...Message)
	Published() []Message
	MsgWriter() io.Writer
}

type context struct {
	context   stdContext.Context
	path      string
	pnames    []string
	pvalues   []string
	handler   HandlerFunc
	hub       *MessagesHub
	topic     string
	sessionID string
	messages  []Message
	rendered  *bytes.Buffer
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

func (c *context) Hub() *MessagesHub {
	return c.hub
}

func (c *context) Publish(messages ...Message) {
	c.hub.Broadcast(c.topic, messages...)
}

func (c *context) Published() []Message {
	return c.messages
}

func (c *context) MsgWriter() io.Writer {
	return c.rendered
}

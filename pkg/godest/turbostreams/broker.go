package turbostreams

import (
	"bytes"
	stdContext "context"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
	"github.com/sargassum-world/fluitans/pkg/godest/pubsub"
)

// Logger is a reduced interface for loggers.
type Logger interface {
	Print(i ...interface{})
	Printf(format string, args ...interface{})
	Debug(i ...interface{})
	Debugf(format string, args ...interface{})
	Info(i ...interface{})
	Infof(format string, args ...interface{})
	Warn(i ...interface{})
	Warnf(format string, args ...interface{})
	Error(i ...interface{})
	Errorf(format string, args ...interface{})
	Fatal(i ...interface{})
	Fatalf(format string, args ...interface{})
	Panic(i ...interface{})
	Panicf(format string, args ...interface{})
}

type Router interface {
	PUB(topic string, h HandlerFunc, m ...MiddlewareFunc) *Route
	SUB(topic string, h HandlerFunc, m ...MiddlewareFunc) *Route
	UNSUB(topic string, h HandlerFunc, m ...MiddlewareFunc) *Route
	MSG(topic string, h HandlerFunc, m ...MiddlewareFunc) *Route
}

type Broker struct {
	hub      *MessagesHub
	router   *router
	maxParam *int
	logger   Logger

	middleware []MiddlewareFunc

	// This is not guarded by a mutex because it's only used by a single goroutine
	pubCancellers map[string]stdContext.CancelFunc
	changes       <-chan pubsub.BroadcastingChange
}

func NewBroker(logger Logger) *Broker {
	changes := make(chan pubsub.BroadcastingChange)
	hub := NewMessagesHub(changes)
	b := &Broker{
		hub:           hub,
		changes:       changes,
		pubCancellers: make(map[string]stdContext.CancelFunc),
		maxParam:      new(int),
		logger:        logger,
	}
	b.router = newRouter(b)
	return b
}

func (b *Broker) Hub() *MessagesHub {
	return b.hub
}

// Handler Registration

func (b *Broker) Add(
	method, topic string, handler HandlerFunc, middleware ...MiddlewareFunc,
) *Route {
	// Copied from github.com/labstack/echo's Echo.add method
	name := handlerName(handler)
	router := b.router
	router.Add(method, topic, func(c Context) error {
		h := applyMiddleware(handler, middleware...)
		return h(c)
	})
	r := &Route{
		Method: method,
		Path:   topic,
		Name:   name,
	}
	b.router.routes[method+topic] = r
	return r
}

func (b *Broker) PUB(topic string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return b.Add(MethodPub, topic, h, m...)
}

func (b *Broker) SUB(topic string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return b.Add(MethodSub, topic, h, m...)
}

func (b *Broker) UNSUB(topic string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return b.Add(MethodUnsub, topic, h, m...)
}

func (b *Broker) MSG(topic string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return b.Add(MethodMsg, topic, h, m...)
}

// Middleware

func (b *Broker) Use(middleware ...MiddlewareFunc) {
	b.middleware = append(b.middleware, middleware...)
}

func (b *Broker) getHandler(method string, topic string, c *context) HandlerFunc {
	b.router.Find(method, topic, c)
	return applyMiddleware(c.handler, b.middleware...)
}

// Action Cable Support

func (b *Broker) ChannelFactory(
	sessionID string, checkers ...actioncable.IdentifierChecker,
) actioncable.ChannelFactory {
	return func(identifier string) (actioncable.Channel, error) {
		return NewChannel(
			identifier, b.hub,
			b.subHandler(sessionID), b.unsubHandler(sessionID), b.msgHandler(sessionID), checkers...,
		)
	}
}

func (b *Broker) newContext(ctx stdContext.Context, topic string) *context {
	return &context{
		context: ctx,
		pvalues: make([]string, *b.maxParam),
		handler: NotFoundHandler,
		hub:     b.hub,
		topic:   topic,
	}
}

func (b *Broker) subHandler(sessionID string) SubHandler {
	return func(ctx stdContext.Context, topic string) error {
		c := b.newContext(ctx, topic)
		c.method = MethodSub
		c.sessionID = sessionID
		h := b.getHandler(MethodSub, topic, c)
		err := errors.Wrapf(h(c), "turbo streams not subscribable on topic %s", topic)
		if err != nil && !errors.Is(err, stdContext.Canceled) {
			b.logger.Error(err)
		}
		return err
	}
}

func (b *Broker) unsubHandler(sessionID string) UnsubHandler {
	return func(ctx stdContext.Context, topic string) {
		c := b.newContext(ctx, topic)
		c.method = MethodUnsub
		c.sessionID = sessionID
		h := b.getHandler(MethodUnsub, topic, c)
		err := errors.Wrapf(h(c), "turbo streams not unsubscribable on topic %s", topic)
		if err != nil && !errors.Is(err, stdContext.Canceled) {
			b.logger.Error(err)
		}
	}
}

func (b *Broker) msgHandler(sessionID string) MsgHandler {
	return func(ctx stdContext.Context, topic string, messages []Message) (result string, err error) {
		c := b.newContext(ctx, topic)
		c.method = MethodMsg
		c.sessionID = sessionID
		c.messages = messages
		c.rendered = &bytes.Buffer{}
		h := b.getHandler(MethodMsg, topic, c)
		err = errors.Wrapf(h(c), "turbo streams message not processable on topic %s", topic)
		if err != nil && !errors.Is(err, stdContext.Canceled) {
			b.logger.Error(err)
			return "", err
		}
		return c.rendered.String(), nil
	}
}

// Managed Publishing

func (b *Broker) startPub(ctx stdContext.Context, topic string) {
	ctx, canceler := stdContext.WithCancel(ctx)
	c := b.newContext(ctx, topic)
	c.method = MethodPub
	b.pubCancellers[topic] = canceler
	h := b.getHandler(MethodPub, topic, c)
	go func() {
		err := h(c)
		if err != nil && !errors.Is(err, stdContext.Canceled) {
			b.logger.Error(err)
		}
	}()
}

func (b *Broker) cancelPub(topic string) {
	if canceller, ok := b.pubCancellers[topic]; ok {
		canceller()
		delete(b.pubCancellers, topic)
	}
}

func (b *Broker) Serve(ctx stdContext.Context) error {
	go func() {
		<-ctx.Done()
		b.hub.Close()
	}()
	for change := range b.changes {
		for _, topic := range change.Added {
			if _, ok := b.pubCancellers[topic]; ok {
				continue
			}
			b.startPub(ctx, topic)
		}
		for _, topic := range change.Removed {
			b.cancelPub(topic)
		}
	}
	return ctx.Err()
}

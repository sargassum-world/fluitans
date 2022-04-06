package turbostreams

import (
	stdContext "context"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
	"github.com/sargassum-world/fluitans/pkg/godest/pubsub"
)

// Interfaces

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
	MSG(topic string, h HandlerFunc, m ...MiddlewareFunc) *Route
}

// Broker

type Broker struct {
	hub           *pubsub.StringHub
	changes       <-chan pubsub.BroadcastingChange
	router        *router
	pubCancellers map[string]stdContext.CancelFunc
	maxParam      *int
	logger        Logger
}

func NewBroker(logger Logger) *Broker {
	changes := make(chan pubsub.BroadcastingChange)
	hub := pubsub.NewStringHub(changes)
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

func (b *Broker) Hub() *pubsub.StringHub {
	return b.hub
}

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

func (b *Broker) MSG(topic string, h HandlerFunc, m ...MiddlewareFunc) *Route {
	return b.Add(MethodMsg, topic, h, m...)
}

func (b *Broker) ChannelFactory(sessionID string, signer Signer) actioncable.ChannelFactory {
	return signer.ChannelFactory(b.hub, b.handleSubscribe(sessionID), b.handleMessage(sessionID))
}

func (b *Broker) newPubContext(topic string) (*context, func()) {
	ctx, canceler := stdContext.WithCancel(stdContext.Background())
	return &context{
		context: ctx,
		pvalues: make([]string, *b.maxParam),
		handler: NotFoundHandler,
		hub:     b.hub,
		topic:   topic,
	}, canceler
}

func (b *Broker) launchPub(topic string) {
	c, canceler := b.newPubContext(topic)
	b.pubCancellers[topic] = canceler
	b.router.Find(MethodPub, topic, c)
	go func() {
		if err := c.handler(c); err != nil && err != stdContext.Canceled {
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

func (b *Broker) newSubContext(topic string, sessionID string) *context {
	return &context{
		context:   stdContext.Background(),
		pvalues:   make([]string, *b.maxParam),
		handler:   NotFoundHandler,
		hub:       b.hub,
		topic:     topic,
		sessionID: sessionID,
	}
}

func (b *Broker) handleSubscribe(sessionID string) SubscribeHandler {
	return func(topic string) error {
		c := b.newSubContext(topic, sessionID)
		b.router.Find(MethodSub, topic, c)
		err := errors.Wrapf(c.handler(c), "turbo streams not subscribable on topic %s", topic)
		if err != nil {
			b.logger.Error(err)
		}
		return err
	}
}

func (b *Broker) handleMessage(sessionID string) MessageHandler {
	return func(topic string, message string) (result string, err error) {
		c := b.newSubContext(topic, sessionID)
		c.message = message
		b.router.Find(MethodMsg, topic, c)
		err = errors.Wrapf(c.handler(c), "turbo streams message not processable on topic %s", topic)
		if err != nil {
			b.logger.Error(err)
			return "", err
		}
		return c.message, nil
	}
}

func (b *Broker) Serve() error {
	// TODO: make this cancellable on a context, and cancel all pubs after the context is done
	for change := range b.changes {
		for _, topic := range change.Added {
			if _, ok := b.pubCancellers[topic]; ok {
				continue
			}
			b.launchPub(topic)
		}
		for _, topic := range change.Removed {
			b.cancelPub(topic)
		}
	}
	return nil
}

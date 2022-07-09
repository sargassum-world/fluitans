// Package actioncable provides a server-side implementation of the Rails Action Cable protocol
// (https://docs.anycable.io/misc/action_cable_protocol)
package actioncable

import (
	"context"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const subprotocol = "actioncable-v1-json"

func Subprotocols() []string {
	return []string{subprotocol}
}

// Messages

type serverMessage struct {
	Type       string `json:"type,omitempty"`
	Identifier string `json:"identifier,omitempty"`
	Message    string `json:"message,omitempty"`
}

type clientMessage struct {
	Command    string `json:"command"`
	Identifier string `json:"identifier"`
	Data       string `json:"data,omitempty"`
}

func newWelcome() serverMessage {
	return serverMessage{
		Type: "welcome",
	}
}

func newPing(t time.Time) serverMessage {
	return serverMessage{
		Type:    "ping",
		Message: fmt.Sprintf("%d", t.Unix()),
	}
}

func newSubscriptionConfirmation(identifier string) serverMessage {
	return serverMessage{
		Type:       "confirm_subscription",
		Identifier: identifier,
	}
}

func newSubscriptionRejection(identifier string) serverMessage {
	return serverMessage{
		Type:       "reject_subscription",
		Identifier: identifier,
	}
}

func newData(identifier, message string) serverMessage {
	return serverMessage{
		Identifier: identifier,
		Message:    message,
	}
}

// Error Handling

func isNormalClose(err error) bool {
	return websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway)
}

func filterNormalClose(underlying error, wrapped error) error {
	if isNormalClose(underlying) {
		// Return the raw error so the Serve function can act differently on a normal close
		return underlying
	}
	return wrapped
}

// Connection

type (
	SubscriptionHandler func(ctx context.Context, sub Subscription) (unsubscriber func(), err error)
	ActionHandler       func(ctx context.Context, identifier, data string) error
	ErrorSanitizer      func(err error) string
)

type Conn struct {
	wsc           *websocket.Conn
	toClient      chan serverMessage
	sh            SubscriptionHandler
	ah            ActionHandler
	unsubscribers map[string]func()
	sanitizeError ErrorSanitizer
}

type ConnOption func(conn *Conn)

func WithSubscriptionHandler(h SubscriptionHandler) ConnOption {
	return func(conn *Conn) {
		conn.sh = h
	}
}

func WithActionHandler(h ActionHandler) ConnOption {
	return func(conn *Conn) {
		conn.ah = h
	}
}

func WithErrorSanitizer(f ErrorSanitizer) ConnOption {
	return func(conn *Conn) {
		conn.sanitizeError = f
	}
}

func defaultErrorSanitizer(err error) string {
	if err == nil {
		return ""
	}

	if errors.Is(err, context.Canceled) {
		return "logged out"
	}
	// Sanitize the error message to avoid leaking information from Serve method errors
	return "server or client error"
}

func Upgrade(wsc *websocket.Conn, opts ...ConnOption) (conn *Conn) {
	conn = &Conn{
		wsc:      wsc,
		toClient: make(chan serverMessage),
		sh: func(ctx context.Context, _ Subscription) (func(), error) {
			return nil, errors.New("No subscription handler registered")
		},
		ah: func(ctx context.Context, _, _ string) error {
			return errors.New("No action handler registered")
		},
		sanitizeError: defaultErrorSanitizer,
		unsubscribers: make(map[string]func()),
	}
	for _, opt := range opts {
		opt(conn)
	}
	return conn
}

func (c *Conn) disconnect(serr error, allowReconnect bool) {
	// Because c.unsubscribers isn't protected by a mutex, the Close method should only be called
	// after the Serve method has completed.
	for _, unsubscriber := range c.unsubscribers {
		unsubscriber()
	}
	c.unsubscribers = make(map[string]func())
	// We leave the toClient channel open because Subscriptions can send into it, and the sendAll
	// method doesn't need to detect whether toClient is closed; Subscriptions can just detect that
	// the connection is done if there's no receiver on the channel.

	// We send close messages only as a courtesy; they may fail if the client already closed the
	// websocket connection by going away, so we don't care about such errors; we need to call the
	// websocket's Close method regardless.
	_ = c.sendJSON(newDisconnect(c.sanitizeError(serr), allowReconnect))
}

func (c *Conn) Close(err error) error {
	// We send close messages only as a courtesy; they may fail if the client already closed the
	// websocket connection by going away, so we don't care about such errors; we need to call the
	// websocket's Close method regardless.
	// TODO: is there any situation where we want to allow reconnection?
	c.disconnect(err, false)
	_ = c.sendMessage(websocket.CloseMessage, []byte{})

	return errors.Wrap(c.wsc.Close(), "couldn't close websocket")
}

// Receiving

const (
	wsPongWait = 60 * time.Second

	subscribeCommand   = "subscribe"
	unsubscribeCommand = "unsubscribe"
	actionCommand      = "message"
)

func (c *Conn) subscribe(ctx context.Context, identifier string) error {
	if _, ok := c.unsubscribers[identifier]; ok {
		c.toClient <- newSubscriptionConfirmation(identifier)
		return nil
	}

	unsubscriber, err := c.sh(ctx, Subscription{
		identifier: identifier,
		toClient:   c.toClient,
	})
	if err != nil {
		c.toClient <- newSubscriptionRejection(identifier)
		return errors.Wrap(err, "subscribe command handler encountered error")
	}
	if unsubscriber == nil {
		c.toClient <- newSubscriptionRejection(identifier)
		return nil
	}
	c.unsubscribers[identifier] = unsubscriber
	c.toClient <- newSubscriptionConfirmation(identifier)
	return nil
}

func (c *Conn) receive(ctx context.Context, command clientMessage) error {
	switch command.Command {
	default:
		return errors.Errorf("unknown command %s", command.Command)
	case subscribeCommand:
		return c.subscribe(ctx, command.Identifier)
	case unsubscribeCommand:
		unsubscriber, ok := c.unsubscribers[command.Identifier]
		if !ok || unsubscriber == nil {
			return nil
		}
		unsubscriber()
		delete(c.unsubscribers, command.Identifier)
	case actionCommand:
		if err := c.ah(ctx, command.Identifier, command.Data); err != nil {
			return errors.Wrap(err, "action command handler encountered error")
		}
	}
	return nil
}

func (c *Conn) receiveAll(ctx context.Context) (err error) {
	if err = c.wsc.SetReadDeadline(time.Now().Add(wsPongWait)); err != nil {
		return errors.Wrap(err, "couldn't set read deadline")
	}
	c.wsc.SetPongHandler(func(string) error {
		return errors.Wrap(
			c.wsc.SetReadDeadline(time.Now().Add(wsPongWait)), "couldn't set read deadline",
		)
	})

	for {
		var command clientMessage
		received := make(chan interface{})
		go func() {
			// ReadJSON blocks for a while due to websocket read timeout, but we don't want it to delay
			// context cancelation so we launch it and synchronize with a closable channel
			err = c.wsc.ReadJSON(&command)
			close(received)
		}()
		select {
		case <-ctx.Done():
			// We wait for received to be closed to avoid a data race where the goroutine for ReadJSON
			// would set err after this select case has already returned an error. We ignore any error
			// from ReadJSON (e.g. broken pipe resulting from browser tab closure, which also cancels the
			// context) because we don't care about reading data after the context is canceled.
			<-received
			return ctx.Err()
		case <-received:
			if err != nil {
				return filterNormalClose(err, errors.Wrap(err, "couldn't parse client message as JSON"))
			}
			if err = ctx.Err(); err != nil {
				return err
			}
			if err = c.receive(ctx, command); err != nil {
				return err
			}
		}
	}
}

// Sending

func (c *Conn) resetWriteDeadline() error {
	const wsWriteWait = 10 * time.Second
	return errors.Wrap(
		c.wsc.SetWriteDeadline(time.Now().Add(wsWriteWait)), "couldn't reset write deadline",
	)
}

func (c *Conn) sendMessage(messageType int, data []byte) error {
	if err := c.resetWriteDeadline(); err != nil {
		return err
	}
	return c.wsc.WriteMessage(messageType, data)
}

func (c *Conn) sendJSON(v interface{}) error {
	if err := c.resetWriteDeadline(); err != nil {
		return err
	}
	return c.wsc.WriteJSON(v)
}

func (c *Conn) sendAll(ctx context.Context) (err error) {
	const (
		wsPingFraction  = 9
		wsPingPeriod    = wsPongWait * wsPingFraction / 10
		cablePingPeriod = 3 * time.Second
	)
	wsPingTicker := time.NewTicker(wsPingPeriod)
	cablePingTicker := time.NewTicker(cablePingPeriod)
	defer func() {
		wsPingTicker.Stop()
		cablePingTicker.Stop()
	}()

	if err = c.resetWriteDeadline(); err != nil {
		return err
	}
	if err = c.wsc.WriteJSON(newWelcome()); err != nil {
		return errors.Wrap(err, "couldn't send welcome message for Action Cable handshake")
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-wsPingTicker.C:
			if err = ctx.Err(); err != nil {
				// Context was also canceled, it should have priority
				return err
			}
			if err = c.sendMessage(
				websocket.PingMessage, []byte(fmt.Sprintf("%d", time.Now().Unix())),
			); err != nil {
				return filterNormalClose(err, errors.Wrap(err, "couldn't send websocket ping"))
			}
		case <-cablePingTicker.C:
			if err = ctx.Err(); err != nil {
				// Context was also canceled, it should have priority
				return err
			}
			if err = c.sendJSON(newPing(time.Now())); err != nil {
				return filterNormalClose(err, errors.Wrap(err, "couldn't send Action Cable ping"))
			}
		case message := <-c.toClient:
			if err = ctx.Err(); err != nil {
				// Context was also canceled, it should have priority
				return err
			}
			if err = c.sendJSON(message); err != nil {
				return filterNormalClose(err, errors.Wrap(
					err, "couldn't send Action Cable server message for client",
				))
			}
		}
	}
}

// Serving

type disconnectMessage struct {
	Type      string `json:"type"`
	Reason    string `json:"reason,omitempty"`
	Reconnect bool   `json:"reconnect,omitempty"`
}

func newDisconnect(reason string, allowReconnect bool) disconnectMessage {
	return disconnectMessage{
		Type:      "disconnect",
		Reason:    reason,
		Reconnect: allowReconnect,
	}
}

func (c *Conn) Serve(ctx context.Context) (err error) {
	eg, egctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		return c.receiveAll(egctx)
	})
	eg.Go(func() error {
		return c.sendAll(egctx)
	})
	if err = eg.Wait(); err != nil {
		if isNormalClose(err) {
			return nil
		}
		return err
	}
	return nil
}

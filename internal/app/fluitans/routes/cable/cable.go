package cable

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/sargassum-world/godest"
	"github.com/sargassum-world/godest/actioncable"
	"github.com/sargassum-world/godest/handling"
	"github.com/sargassum-world/godest/session"
	"github.com/sargassum-world/godest/turbostreams"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
)

func serveWSConn(
	r *http.Request, wsc *websocket.Conn, sess *sessions.Session,
	channelFactories map[string]actioncable.ChannelFactory,
	cc *session.CSRFTokenChecker, acc *actioncable.Cancellers, wsu websocket.Upgrader,
	l godest.Logger,
) {
	conn, err := actioncable.Upgrade(wsc, actioncable.NewChannelDispatcher(
		channelFactories, make(map[string]actioncable.Channel),
		actioncable.WithCSRFTokenChecker(func(token string) error {
			return cc.Check(r, token)
		}),
	))
	if err != nil {
		l.Error(errors.Wrapf(
			err,
			"couldn't upgrade websocket connection to action cable connection "+
				"(client requested subprotocols %v, upgrader supports subprotocols %v)",
			websocket.Subprotocols(r),
			wsu.Subprotocols,
		))
		if cerr := wsc.Close(); cerr != nil {
			l.Error(errors.Wrapf(cerr, "couldn't close websocket"))
		}
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	acc.Add(sess.ID, cancel)
	serr := handling.Except(conn.Serve(ctx), context.Canceled)
	if serr != nil {
		// We can't return errors after the HTTP request is upgraded to a websocket, so we just log them
		l.Error(serr)
	}
	if cerr := conn.Close(serr); err != nil {
		// We can't return errors after the HTTP request is upgraded to a websocket, so we just log them
		l.Error(cerr)
	}
}

func (h *Handlers) HandleCableGet() auth.HTTPHandlerFuncWithSession {
	return func(c echo.Context, _ auth.Auth, sess *sessions.Session) error {
		wsc, err := h.wsu.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return errors.Wrap(err, "couldn't upgrade http request to websocket connection")
		}

		const wsMaxMessageSize = 512
		wsc.SetReadLimit(wsMaxMessageSize)
		serveWSConn(
			c.Request(), wsc, sess,
			map[string]actioncable.ChannelFactory{
				turbostreams.ChannelName: turbostreams.NewChannelFactory(h.tsb, sess.ID, h.acs.Check),
			},
			h.cc, h.acc, h.wsu, h.l,
		)
		return nil
	}
}

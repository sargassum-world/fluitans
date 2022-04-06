package cable

import (
	"context"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"

	"github.com/sargassum-world/fluitans/internal/app/fluitans/auth"
	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
	"github.com/sargassum-world/fluitans/pkg/godest/turbostreams"
)

func (h *Handlers) HandleCableGet() auth.HTTPHandlerFuncWithSession {
	return func(c echo.Context, _ auth.Auth, sess *sessions.Session) error {
		wsc, err := h.wsu.Upgrade(c.Response(), c.Request(), nil)
		if err != nil {
			return err
		}

		const wsMaxMessageSize = 512
		wsc.SetReadLimit(wsMaxMessageSize)

		acc := actioncable.Upgrade(wsc, actioncable.WithChannels(
			map[string]actioncable.ChannelFactory{
				turbostreams.ChannelName: h.tsb.ChannelFactory(sess.ID, h.tss),
			},
			make(map[string]actioncable.Channel),
		))
		ctx, cancel := context.WithCancel(c.Request().Context())
		h.acc.Add(sess.ID, cancel)
		serr := acc.Serve(ctx)
		// We can't return errors after the HTTP request is upgraded to a websocket, so we just log them
		if serr != nil && serr != context.Canceled {
			h.l.Error(serr)
		}
		if err := acc.Close(); err != nil {
			h.l.Error(err)
		}
		return nil
	}
}

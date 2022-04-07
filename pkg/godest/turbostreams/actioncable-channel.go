package turbostreams

import (
	stdContext "context"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
)

const ChannelName = "Turbo::StreamsChannel"

type (
	SubHandler func(ctx stdContext.Context, streamName string) error
	MsgHandler func(
		ctx stdContext.Context, streamName string, messages []Message,
	) (result string, err error)
)

type Channel struct {
	identifier string
	name       signedName
	h          *MessagesHub
	handleSub  SubHandler
	handleMsg  MsgHandler
}

func (c *Channel) Subscribe(
	ctx stdContext.Context, sub actioncable.Subscription,
) (unsubscriber func(), err error) {
	if sub.Identifier() != c.identifier {
		return nil, errors.Errorf(
			"channel identifier %+v does not match subscription identifier %+v",
			c.identifier, sub.Identifier(),
		)
	}
	streamName := c.name.Name
	if err := c.handleSub(ctx, streamName); err != nil {
		return nil, nil // since subscribing isn't possible/authorized, reject the subscription
	}
	ctx, cancel := stdContext.WithCancel(ctx)
	unsub, removed := c.h.Subscribe(streamName, func(messages []Message) (ok bool) {
		if ctx.Err() != nil {
			return false
		}
		result, err := c.handleMsg(ctx, streamName, messages)
		if err != nil {
			cancel()
			sub.Close()
			return false
		}
		return sub.Receive(result)
	})
	go func() {
		select {
		case <-ctx.Done():
			break
		case <-removed:
			break
		}
		cancel()
		unsub()
		sub.Close()
	}()
	return cancel, nil
}

func (c *Channel) Perform(data string) error {
	return errors.New("turbo streams channel cannot perform any actions")
}

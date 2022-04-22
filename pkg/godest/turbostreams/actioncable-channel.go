package turbostreams

import (
	stdContext "context"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
)

const ChannelName = "Turbo::StreamsChannel"

type (
	SubHandler   func(ctx stdContext.Context, streamName string) error
	UnsubHandler func(ctx stdContext.Context, streamName string)
	MsgHandler   func(
		ctx stdContext.Context, streamName string, messages []Message,
	) (result string, err error)
)

type Channel struct {
	identifier  string
	streamName  string
	h           *MessagesHub
	handleSub   SubHandler
	handleUnsub UnsubHandler
	handleMsg   MsgHandler
}

func parseStreamName(identifier string) (string, error) {
	var i struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal([]byte(identifier), &i); err != nil {
		return "", errors.Wrap(err, "couldn't parse stream name from identifier")
	}
	return i.Name, nil
}

func NewChannel(
	identifier string, h *MessagesHub,
	handleSub SubHandler, handleUnsub UnsubHandler, handleMsg MsgHandler,
	checkers ...actioncable.IdentifierChecker,
) (*Channel, error) {
	name, err := parseStreamName(identifier)
	if err != nil {
		return nil, err
	}
	for _, checker := range checkers {
		if err := checker(identifier); err != nil {
			return nil, errors.Wrap(err, "stream identifier failed checks")
		}
	}
	return &Channel{
		identifier:  identifier,
		streamName:  name,
		h:           h,
		handleSub:   handleSub,
		handleUnsub: handleUnsub,
		handleMsg:   handleMsg,
	}, nil
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
	if err := c.handleSub(ctx, c.streamName); err != nil {
		return nil, nil // since subscribing isn't possible/authorized, reject the subscription
	}
	ctx, cancel := stdContext.WithCancel(ctx)
	unsub, removed := c.h.Subscribe(c.streamName, func(messages []Message) (ok bool) {
		if ctx.Err() != nil {
			return false
		}
		result, err := c.handleMsg(ctx, c.streamName, messages)
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
		c.handleUnsub(ctx, c.streamName)
		sub.Close()
	}()
	return cancel, nil
}

func (c *Channel) Perform(data string) error {
	return errors.New("turbo streams channel cannot perform any actions")
}

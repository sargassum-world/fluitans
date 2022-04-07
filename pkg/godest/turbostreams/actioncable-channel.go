package turbostreams

import (
	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
)

const ChannelName = "Turbo::StreamsChannel"

type (
	SubHandler func(streamName string) error
	MsgHandler func(streamName string, messages []Message) (result string, err error)
)

type Channel struct {
	identifier string
	name       signedName
	h          *MessagesHub
	handleSub  SubHandler
	handleMsg  MsgHandler
}

func (c *Channel) Subscribe(sub actioncable.Subscription) (unsubscriber func(), err error) {
	if sub.Identifier() != c.identifier {
		return nil, errors.Errorf(
			"channel identifier %+v does not match subscription identifier %+v",
			c.identifier, sub.Identifier(),
		)
	}
	streamName := c.name.Name
	if err := c.handleSub(streamName); err != nil {
		return nil, nil // since subscribing isn't possible/authorized, reject the subscription
	}
	// TODO: subscription should be to interface{}, so that handleMsg transforms it into a string
	return c.h.Subscribe(streamName, func(messages []Message) (ok bool) {
		result, err := c.handleMsg(streamName, messages)
		if err != nil {
			// Since receiving isn't possible/authorized, cancel subscriptions to the topic
			sub.Close()
			return false
		}
		return sub.Receive(result)
	}), nil
}

func (c *Channel) Perform(data string) error {
	return errors.New("turbo streams channel cannot perform any actions")
}

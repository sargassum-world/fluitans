package turbostreams

import (
	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
	"github.com/sargassum-world/fluitans/pkg/godest/pubsub"
)

const ChannelName = "Turbo::StreamsChannel"

type (
	SubscribeHandler func(streamName string) error
	MessageHandler   func(streamName, data string) (result string, err error)
)

type Channel struct {
	identifier      string
	name            signedName
	h               *pubsub.StringHub
	handleSubscribe SubscribeHandler
	handleMessage   MessageHandler
}

func (c *Channel) Subscribe(sub actioncable.Subscription) (unsubscriber func(), err error) {
	if sub.Identifier() != c.identifier {
		return nil, errors.Errorf(
			"channel identifier %+v does not match subscription identifier %+v",
			c.identifier, sub.Identifier(),
		)
	}
	streamName := c.name.Name
	if err := c.handleSubscribe(streamName); err != nil {
		return nil, nil // since subscribing isn't possible/authorized, reject the subscription
	}
	// TODO: subscription should be to interface{}, so that handleMessage transforms it into a string
	return c.h.Subscribe(streamName, func(message string) (ok bool) {
		result, err := c.handleMessage(streamName, message)
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

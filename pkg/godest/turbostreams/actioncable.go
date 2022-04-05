package turbostreams

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/sargassum-world/fluitans/pkg/godest/actioncable"
	"github.com/sargassum-world/fluitans/pkg/godest/pubsub"
)

const ChannelName = "Turbo::StreamsChannel"

type signedName struct {
	Name string `json:"name"`
	HMAC string `json:"hmac"`
}

func parseIdentifier(identifier string) (signedName, error) {
	var p struct {
		SignedName string `json:"signed_stream_name"`
	}
	if err := json.Unmarshal([]byte(identifier), &p); err != nil {
		return signedName{}, errors.Wrap(err, "couldn't parse identifier for params")
	}
	name := signedName{
		Name: p.SignedName,
		// TODO: fix this!
	}
	// TODO: check the signature
	return name, nil
}

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

func NewChannel(
	identifier string, h *pubsub.StringHub,
	handleSubscribe SubscribeHandler, handleMessage MessageHandler,
) (*Channel, error) {
	name, err := parseIdentifier(identifier)
	if err != nil {
		return nil, err
	}
	return &Channel{
		identifier:      identifier,
		name:            name,
		h:               h,
		handleSubscribe: handleSubscribe,
		handleMessage:   handleMessage,
	}, nil
}

func ChannelFactory(
	h *pubsub.StringHub, handleSubscribe SubscribeHandler, handleMessage MessageHandler,
) actioncable.ChannelFactory {
	return func(identifier string) (actioncable.Channel, error) {
		return NewChannel(identifier, h, handleSubscribe, handleMessage)
	}
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

package actioncable

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
)

// Channel

func ParseIdentifier(identifierRaw string) (channelName string, err error) {
	var i struct {
		Channel string `json:"channel"`
	}
	if err := json.Unmarshal([]byte(identifierRaw), &i); err != nil {
		return "", errors.Wrap(err, "couldn't parse identifier")
	}
	return i.Channel, nil
}

type Channel interface {
	Subscribe(ctx context.Context, sub Subscription) (unsubscriber func(), err error)
	Perform(data string) error
}

type ChannelFactory func(identifier string) (Channel, error)

func HandleSubscription(
	factories map[string]ChannelFactory, channels map[string]Channel,
) SubscriptionHandler {
	return func(ctx context.Context, sub Subscription) (unsubscriber func(), err error) {
		if channel, ok := channels[sub.Identifier()]; ok {
			unsubscriber, err = channel.Subscribe(ctx, sub)
			if unsubscriber == nil || err != nil {
				delete(channels, sub.Identifier())
				return nil, errors.Wrapf(err, "couldn't re-subscribe to %s", sub.Identifier())
			}
			return unsubscriber, nil
		}

		channelName, err := ParseIdentifier(sub.Identifier())
		if err != nil {
			return nil, err
		}
		factory, ok := factories[channelName]
		if !ok {
			return nil, errors.Errorf("unknown channel name %s", channelName)
		}
		channel, err := factory(sub.Identifier())
		if err != nil {
			return nil, errors.Wrapf(err, "couldn't instantiate %s", sub.Identifier())
		}
		// The Subscribe method checks whether the subscription is possible, so we must check it before
		// storing the channel.
		unsubscriber, err = channel.Subscribe(ctx, sub)
		if unsubscriber == nil || err != nil {
			return nil, errors.Wrapf(err, "couldn't subscribe to %s", sub.Identifier())
		}
		channels[sub.Identifier()] = channel
		return unsubscriber, nil
	}
}

func HandleAction(channels map[string]Channel) ActionHandler {
	return func(ctx context.Context, identifier, data string) error {
		channel, ok := channels[identifier]
		if !ok {
			return errors.Errorf("no preexisting subscription on %s", identifier)
		}
		return channel.Perform(data)
	}
}

func WithChannels(
	factories map[string]ChannelFactory, channels map[string]Channel,
) ConnOption {
	return func(conn *Conn) {
		conn.sh = HandleSubscription(factories, channels)
		conn.ah = HandleAction(channels)
	}
}

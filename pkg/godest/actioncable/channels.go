package actioncable

import (
	"context"

	"github.com/pkg/errors"
)

type Channel interface {
	Subscribe(ctx context.Context, sub Subscription) (unsubscriber func(), err error)
	Perform(data string) error
}

type ChannelFactory func(identifier string) (Channel, error)

func HandleSubscription(
	factories map[string]ChannelFactory, channels map[string]Channel, checkers ...IdentifierChecker,
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

		channelName, err := parseChannelName(sub.Identifier())
		if err != nil {
			return nil, err
		}
		for _, checker := range checkers {
			if err = checker(sub.Identifier()); err != nil {
				return nil, errors.Wrap(err, "channel identifier failed check")
			}
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
	factories map[string]ChannelFactory, channels map[string]Channel, checkers ...IdentifierChecker,
) ConnOption {
	return func(conn *Conn) {
		conn.sh = HandleSubscription(factories, channels, checkers...)
		conn.ah = HandleAction(channels)
	}
}

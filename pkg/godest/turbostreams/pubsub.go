package turbostreams

import (
	"github.com/sargassum-world/fluitans/pkg/godest/pubsub"
)

type MessagesReceiveFunc func(messages []Message) (ok bool)

type MessagesHub struct {
	hub *pubsub.DataHub
}

func NewMessagesHub(brChanges chan<- pubsub.BroadcastingChange) *MessagesHub {
	return &MessagesHub{hub: pubsub.NewDataHub(brChanges)}
}

func (h *MessagesHub) Subscribe(
	topic string, receive MessagesReceiveFunc,
) (unsubscriber func(), removed <-chan struct{}) {
	return h.hub.Subscribe(
		topic,
		func(message interface{}) (ok bool) {
			m, ok := message.([]Message)
			if !ok {
				// This should never happen because Broadcast only takes Messages
				return false
			}
			return receive(m)
		},
	)
}

func (h *MessagesHub) Cancel(topics ...string) {
	h.hub.Cancel(topics...)
}

func (h *MessagesHub) Broadcast(topic string, messages ...Message) {
	h.hub.Broadcast(topic, messages)
}

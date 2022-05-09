package pubsub

type StringReceiveFunc func(message string) (ok bool)

type StringHub struct {
	hub *DataHub
}

func NewStringHub(brChanges chan<- BroadcastingChange) *StringHub {
	return &StringHub{hub: NewDataHub(brChanges)}
}

func (h *StringHub) Shutdown() {
	h.hub.Shutdown()
}

func (h *StringHub) Subscribe(
	topic string, receive StringReceiveFunc,
) (unsubscriber func(), removed <-chan struct{}) {
	return h.hub.Subscribe(
		topic,
		func(message interface{}) (ok bool) {
			m, ok := message.(string)
			if !ok {
				// This should never happen because Broadcast only takes strings
				return false
			}
			return receive(m)
		},
	)
}

func (h *StringHub) Cancel(topics ...string) {
	h.hub.Cancel(topics...)
}

func (h *StringHub) Broadcast(topic, message string) {
	h.hub.Broadcast(topic, message)
}

package pubsub

type StringReceiveFunc func(message string) (ok bool)

type StringHub struct {
	hub *DataHub
}

func NewStringHub(brChanges chan<- BroadcastingChange) *StringHub {
	return &StringHub{hub: NewDataHub(brChanges)}
}

func (h *StringHub) Subscribe(topic string, receive StringReceiveFunc) (unsubscriber func()) {
	return h.hub.Subscribe(topic, func(message interface{}) (ok bool) {
		m, ok := message.(string)
		if !ok {
			// This should never happen because Broadcast only takes strings
			return false
		}
		return receive(m)
	})
}

func (h *StringHub) Broadcast(topic, message string) {
	h.hub.Broadcast(topic, message)
}

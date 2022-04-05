// Package pubsub provides pub-sub functionality.
package pubsub

import (
	"sync"

	"github.com/pkg/errors"
)

type DataReceiveFunc func(message interface{}) (ok bool)

type dataReceiver struct {
	topic   string
	receive DataReceiveFunc
}

type DataHub struct {
	receivers map[string]map[*dataReceiver]bool
	mu        sync.RWMutex
	brChanges chan<- BroadcastingChange
}

func NewDataHub(brChanges chan<- BroadcastingChange) *DataHub {
	return &DataHub{
		receivers: make(map[string]map[*dataReceiver]bool),
		brChanges: brChanges,
	}
}

func (h *DataHub) Subscribe(topic string, receive DataReceiveFunc) (unsubscriber func()) {
	receiver := &dataReceiver{
		topic:   topic,
		receive: receive,
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	broadcasting, ok := h.receivers[topic]
	addedTopic := false
	if !ok {
		broadcasting = make(map[*dataReceiver]bool)
		h.receivers[receiver.topic] = broadcasting
		addedTopic = true
	}
	broadcasting[receiver] = true

	if h.brChanges != nil && addedTopic {
		h.brChanges <- BroadcastingChange{Added: []string{topic}}
	}

	return func() {
		h.unsubscribe(receiver)
	}
}

func (h *DataHub) unsubscribe(receivers ...*dataReceiver) {
	if len(receivers) == 0 {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	removedTopics := make([]string, 0, len(receivers))
	for _, receiver := range receivers {
		broadcasting, ok := h.receivers[receiver.topic]
		if !ok {
			continue
		}
		delete(broadcasting, receiver)
		if len(broadcasting) == 0 {
			delete(h.receivers, receiver.topic)
			removedTopics = append(removedTopics, receiver.topic)
		}
	}
	if h.brChanges != nil && len(removedTopics) > 0 {
		h.brChanges <- BroadcastingChange{Removed: removedTopics}
	}
}

func (h *DataHub) Broadcast(topic string, message interface{}) {
	unsubscribe, _ := h.broadcastStrict(topic, message)
	h.unsubscribe(unsubscribe...)
}

func (h *DataHub) broadcastStrict(
	topic string, message interface{},
) (unsubscribe []*dataReceiver, err error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	broadcasting, ok := h.receivers[topic]
	if !ok {
		return nil, errors.Errorf("no receivers for %s", topic)
	}
	willUnsubscribe := make(chan *dataReceiver, len(broadcasting))
	wg := sync.WaitGroup{}
	for receiver := range broadcasting {
		wg.Add(1)
		go func(receiver *dataReceiver, message interface{}, willUnsubscribe chan<- *dataReceiver) {
			defer wg.Done()
			if !receiver.receive(message) {
				willUnsubscribe <- receiver
				return
			}
		}(receiver, message, willUnsubscribe)
	}
	wg.Wait()

	close(willUnsubscribe)
	unsubscribe = make([]*dataReceiver, 0, len(broadcasting))
	for receiver := range willUnsubscribe {
		unsubscribe = append(unsubscribe, receiver)
	}
	return unsubscribe, nil
}

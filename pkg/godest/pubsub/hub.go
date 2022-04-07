// Package pubsub provides pub-sub functionality.
package pubsub

import (
	"context"
	"sync"

	"github.com/pkg/errors"
)

// Receiver

type DataReceiveFunc func(message interface{}) (ok bool)

type dataReceiver struct {
	topic   string
	receive DataReceiveFunc
	cancel  context.CancelFunc
}

// Hub

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

func (h *DataHub) Subscribe(
	topic string, receive DataReceiveFunc,
) (unsubscriber func(), removed <-chan struct{}) {
	ctx, cancel := context.WithCancel(context.Background())
	receiver := &dataReceiver{
		topic:   topic,
		receive: receive,
		cancel:  cancel,
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
		cancel()
		h.unsubscribe(receiver)
	}, ctx.Done()
}

func (h *DataHub) Cancel(topics ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, topic := range topics {
		broadcasting, ok := h.receivers[topic]
		if !ok {
			continue
		}
		for receiver := range broadcasting {
			receiver.cancel()
		}
		delete(h.receivers, topic)
	}

	if h.brChanges != nil && len(topics) > 0 {
		h.brChanges <- BroadcastingChange{Removed: topics}
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
		receiver.cancel()
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

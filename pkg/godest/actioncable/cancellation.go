package actioncable

import (
	"context"
	"sync"
)

type Cancellers struct {
	funcs map[string][]context.CancelFunc
	m     *sync.Mutex
}

func NewCancellers() *Cancellers {
	return &Cancellers{
		funcs: make(map[string][]context.CancelFunc),
		m:     &sync.Mutex{},
	}
}

func (c *Cancellers) Add(id string, canceller context.CancelFunc) {
	// Note: To reduce lock contention at larger scales, we could have a map associating each id to
	// a mutex for the value in the funcs map; but when the map of mutexes is missing a key, we'd need
	// to add an entry. This extra complexity may be justified in the future.
	c.m.Lock()
	defer c.m.Unlock()

	c.funcs[id] = append(c.funcs[id], canceller)
}

func (c *Cancellers) Cancel(id string) {
	c.m.Lock()
	defer c.m.Unlock()

	for _, canceller := range c.funcs[id] {
		canceller()
	}
	delete(c.funcs, id)
}

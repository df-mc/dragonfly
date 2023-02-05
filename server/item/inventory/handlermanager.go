package inventory

import (
	"sync"
)

// HandlerManager manages a inventory's handlers.
type HandlerManager struct {
	sync.Mutex
	handlers []Handler
}

type Handler interface {
	HandleDrop(EventDrop)
	HandlePlace(EventPlace)
	HandleTake(EventTake)
}

func (hm *HandlerManager) AddHandler(h Handler) func(Handler) Handler {
	hm.Lock()
	defer hm.Unlock()

	idx := len(hm.handlers)
	hm.handlers = append(hm.handlers, h)

	return func(hNew Handler) Handler {
		hm.Lock()
		defer hm.Unlock()

		if hNew == nil {
			hm.handlers[idx] = NopHandler{}
			return h
		}

		hm.handlers[idx] = hNew
		return h
	}
}

func (hm *HandlerManager) HandleDrop(e EventDrop) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleDrop(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandlePlace(e EventPlace) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandlePlace(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleTake(e EventTake) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleTake(e)
	}

	hm.Unlock()
}

type NopHandler struct{}

func (NopHandler) HandleDrop(EventDrop)   {}
func (NopHandler) HandlePlace(EventPlace) {}
func (NopHandler) HandleTake(EventTake)   {}

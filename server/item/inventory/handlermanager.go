package inventory

import (
	"github.com/df-mc/atomic"
	"sync"
)

// HandlerManager manages a inventory's handlers.
type HandlerManager struct {
	sync.Mutex
	handlers []*atomic.Value[Handler]
}

type Handler interface {
	HandleDrop(EventDrop)
	HandlePlace(EventPlace)
	HandleTake(EventTake)
}

func (hm *HandlerManager) AddHandler(h Handler) func(Handler) Handler {
	hm.Lock()
	defer hm.Unlock()

	ah := atomic.NewValue[Handler](h)
	hm.handlers = append(hm.handlers, ah)

	return func(newHandler Handler) Handler {
		hm.Lock()
		defer hm.Unlock()

		if newHandler == nil {
			return ah.Swap(NopHandler{})
		}

		return ah.Swap(newHandler)
	}
}

func (hm *HandlerManager) HandleDrop(e EventDrop) {
	for _, h := range hm.handlers {
		h.Load().HandleDrop(e)
	}
}

func (hm *HandlerManager) HandlePlace(e EventPlace) {
	for _, h := range hm.handlers {
		h.Load().HandlePlace(e)
	}
}

func (hm *HandlerManager) HandleTake(e EventTake) {
	for _, h := range hm.handlers {
		h.Load().HandleTake(e)
	}
}

type NopHandler struct{}

func (NopHandler) HandleDrop(EventDrop)   {}
func (NopHandler) HandlePlace(EventPlace) {}
func (NopHandler) HandleTake(EventTake)   {}

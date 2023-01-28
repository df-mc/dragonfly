package world

import (
	"sync"
)

// HandlerManager manages a world's handlers.
type HandlerManager struct {
	sync.Mutex
	handlers []Handler
}

type Handler interface {
	HandleBlockBurn(EventBlockBurn)
	HandleClose(EventClose)
	HandleEntityDespawn(EventEntityDespawn)
	HandleEntitySpawn(EventEntitySpawn)
	HandleFireSpread(EventFireSpread)
	HandleLiquidDecay(EventLiquidDecay)
	HandleLiquidFlow(EventLiquidFlow)
	HandleLiquidHarden(EventLiquidHarden)
	HandleSound(EventSound)
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

func (hm *HandlerManager) HandleBlockBurn(e EventBlockBurn) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleBlockBurn(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleClose(e EventClose) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleClose(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleEntityDespawn(e EventEntityDespawn) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleEntityDespawn(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleEntitySpawn(e EventEntitySpawn) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleEntitySpawn(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleFireSpread(e EventFireSpread) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleFireSpread(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleLiquidDecay(e EventLiquidDecay) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleLiquidDecay(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleLiquidFlow(e EventLiquidFlow) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleLiquidFlow(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleLiquidHarden(e EventLiquidHarden) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleLiquidHarden(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleSound(e EventSound) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleSound(e)
	}

	hm.Unlock()
}

type NopHandler struct{}

func (NopHandler) HandleBlockBurn(EventBlockBurn)         {}
func (NopHandler) HandleClose(EventClose)                 {}
func (NopHandler) HandleEntityDespawn(EventEntityDespawn) {}
func (NopHandler) HandleEntitySpawn(EventEntitySpawn)     {}
func (NopHandler) HandleFireSpread(EventFireSpread)       {}
func (NopHandler) HandleLiquidDecay(EventLiquidDecay)     {}
func (NopHandler) HandleLiquidFlow(EventLiquidFlow)       {}
func (NopHandler) HandleLiquidHarden(EventLiquidHarden)   {}
func (NopHandler) HandleSound(EventSound)                 {}

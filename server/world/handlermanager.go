package world

import (
	"github.com/df-mc/atomic"
	"sync"
)

// HandlerManager manages a world's handlers.
type HandlerManager struct {
	sync.Mutex
	handlers []*atomic.Value[Handler]
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

func (hm *HandlerManager) HandleBlockBurn(e EventBlockBurn) {
	for _, h := range hm.handlers {
		h.Load().HandleBlockBurn(e)
	}
}

func (hm *HandlerManager) HandleClose(e EventClose) {
	for _, h := range hm.handlers {
		h.Load().HandleClose(e)
	}
}

func (hm *HandlerManager) HandleEntityDespawn(e EventEntityDespawn) {
	for _, h := range hm.handlers {
		h.Load().HandleEntityDespawn(e)
	}
}

func (hm *HandlerManager) HandleEntitySpawn(e EventEntitySpawn) {
	for _, h := range hm.handlers {
		h.Load().HandleEntitySpawn(e)
	}
}

func (hm *HandlerManager) HandleFireSpread(e EventFireSpread) {
	for _, h := range hm.handlers {
		h.Load().HandleFireSpread(e)
	}
}

func (hm *HandlerManager) HandleLiquidDecay(e EventLiquidDecay) {
	for _, h := range hm.handlers {
		h.Load().HandleLiquidDecay(e)
	}
}

func (hm *HandlerManager) HandleLiquidFlow(e EventLiquidFlow) {
	for _, h := range hm.handlers {
		h.Load().HandleLiquidFlow(e)
	}
}

func (hm *HandlerManager) HandleLiquidHarden(e EventLiquidHarden) {
	for _, h := range hm.handlers {
		h.Load().HandleLiquidHarden(e)
	}
}

func (hm *HandlerManager) HandleSound(e EventSound) {
	for _, h := range hm.handlers {
		h.Load().HandleSound(e)
	}
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

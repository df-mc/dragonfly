package player

import (
	"github.com/df-mc/atomic"
	"sync"
)

// HandlerManager manages a player's handlers.
type HandlerManager struct {
	sync.Mutex
	handlers map[int]atomic.Value[Handler]
}

type Handler interface {
	HandleAttackEntity(EventAttackEntity)
	HandleBlockBreak(EventBlockBreak)
	HandleBlockPick(EventBlockPick)
	HandleBlockPlace(EventBlockPlace)
	HandleChangeWorld(EventChangeWorld)
	HandleChat(EventChat)
	HandleCommandExecution(EventCommandExecution)
	HandleDeath(EventDeath)
	HandleExperienceGain(EventExperienceGain)
	HandleFoodLoss(EventFoodLoss)
	HandleHeal(EventHeal)
	HandleHurt(EventHurt)
	HandleItemConsume(EventItemConsume)
	HandleItemDamage(EventItemDamage)
	HandleItemDrop(EventItemDrop)
	HandleItemPickup(EventItemPickup)
	HandleItemUse(EventItemUse)
	HandleItemUseOnBlock(EventItemUseOnBlock)
	HandleItemUseOnEntity(EventItemUseOnEntity)
	HandleJump(EventJump)
	HandleMove(EventMove)
	HandlePunchAir(EventPunchAir)
	HandleQuit(EventQuit)
	HandleRespawn(EventRespawn)
	HandleSignEdit(EventSignEdit)
	HandleSkinChange(EventSkinChange)
	HandleStartBreak(EventStartBreak)
	HandleTeleport(EventTeleport)
	HandleToggleSneak(EventToggleSneak)
	HandleToggleSprint(EventToggleSprint)
	HandleTransfer(EventTransfer)
}

func (hm *HandlerManager) AddHandler(h Handler) func(Handler) {
	hm.Lock()
	handlerID := findKey[atomic.Value[Handler]](hm.handlers)
	hm.handlers[handlerID] = *atomic.NewValue[Handler](h)
	hm.Unlock()

	return func(newHandler Handler) {
		if newHandler == nil {
			handler := hm.handlers[handlerID]
			handler.Store(NopHandler{})
			return
		}

		handler := hm.handlers[handlerID]
		handler.Store(newHandler)
	}
}

func (hm *HandlerManager) HandleAttackEntity(e EventAttackEntity) {
	for _, handler := range hm.handlers {
		handler.Load().HandleAttackEntity(e)
	}
}

func (hm *HandlerManager) HandleBlockBreak(e EventBlockBreak) {
	for _, handler := range hm.handlers {
		handler.Load().HandleBlockBreak(e)
	}
}

func (hm *HandlerManager) HandleBlockPick(e EventBlockPick) {
	for _, handler := range hm.handlers {
		handler.Load().HandleBlockPick(e)
	}
}

func (hm *HandlerManager) HandleBlockPlace(e EventBlockPlace) {
	for _, handler := range hm.handlers {
		handler.Load().HandleBlockPlace(e)
	}
}

func (hm *HandlerManager) HandleChangeWorld(e EventChangeWorld) {
	for _, handler := range hm.handlers {
		handler.Load().HandleChangeWorld(e)
	}
}

func (hm *HandlerManager) HandleChat(e EventChat) {
	for _, handler := range hm.handlers {
		handler.Load().HandleChat(e)
	}
}

func (hm *HandlerManager) HandleCommandExecution(e EventCommandExecution) {
	for _, handler := range hm.handlers {
		handler.Load().HandleCommandExecution(e)
	}
}

func (hm *HandlerManager) HandleDeath(e EventDeath) {
	for _, handler := range hm.handlers {
		handler.Load().HandleDeath(e)
	}
}

func (hm *HandlerManager) HandleExperienceGain(e EventExperienceGain) {
	for _, handler := range hm.handlers {
		handler.Load().HandleExperienceGain(e)
	}
}

func (hm *HandlerManager) HandleFoodLoss(e EventFoodLoss) {
	for _, handler := range hm.handlers {
		handler.Load().HandleFoodLoss(e)
	}
}

func (hm *HandlerManager) HandleHeal(e EventHeal) {
	for _, handler := range hm.handlers {
		handler.Load().HandleHeal(e)
	}
}

func (hm *HandlerManager) HandleHurt(e EventHurt) {
	for _, handler := range hm.handlers {
		handler.Load().HandleHurt(e)
	}
}

func (hm *HandlerManager) HandleItemConsume(e EventItemConsume) {
	for _, handler := range hm.handlers {
		handler.Load().HandleItemConsume(e)
	}
}

func (hm *HandlerManager) HandleItemDamage(e EventItemDamage) {
	for _, handler := range hm.handlers {
		handler.Load().HandleItemDamage(e)
	}
}

func (hm *HandlerManager) HandleItemDrop(e EventItemDrop) {
	for _, handler := range hm.handlers {
		handler.Load().HandleItemDrop(e)
	}
}

func (hm *HandlerManager) HandleItemPickup(e EventItemPickup) {
	for _, handler := range hm.handlers {
		handler.Load().HandleItemPickup(e)
	}
}

func (hm *HandlerManager) HandleItemUse(e EventItemUse) {
	for _, handler := range hm.handlers {
		handler.Load().HandleItemUse(e)
	}
}

func (hm *HandlerManager) HandleItemUseOnBlock(e EventItemUseOnBlock) {
	for _, handler := range hm.handlers {
		handler.Load().HandleItemUseOnBlock(e)
	}
}

func (hm *HandlerManager) HandleItemUseOnEntity(e EventItemUseOnEntity) {
	for _, handler := range hm.handlers {
		handler.Load().HandleItemUseOnEntity(e)
	}
}

func (hm *HandlerManager) HandleJump(e EventJump) {
	for _, handler := range hm.handlers {
		handler.Load().HandleJump(e)
	}
}

func (hm *HandlerManager) HandleMove(e EventMove) {
	for _, handler := range hm.handlers {
		handler.Load().HandleMove(e)
	}
}

func (hm *HandlerManager) HandlePunchAir(e EventPunchAir) {
	for _, handler := range hm.handlers {
		handler.Load().HandlePunchAir(e)
	}
}

func (hm *HandlerManager) HandleQuit(e EventQuit) {
	for _, handler := range hm.handlers {
		handler.Load().HandleQuit(e)
	}
}

func (hm *HandlerManager) HandleRespawn(e EventRespawn) {
	for _, handler := range hm.handlers {
		handler.Load().HandleRespawn(e)
	}
}

func (hm *HandlerManager) HandleSignEdit(e EventSignEdit) {
	for _, handler := range hm.handlers {
		handler.Load().HandleSignEdit(e)
	}
}

func (hm *HandlerManager) HandleSkinChange(e EventSkinChange) {
	for _, handler := range hm.handlers {
		handler.Load().HandleSkinChange(e)
	}
}

func (hm *HandlerManager) HandleStartBreak(e EventStartBreak) {
	for _, handler := range hm.handlers {
		handler.Load().HandleStartBreak(e)
	}
}

func (hm *HandlerManager) HandleTeleport(e EventTeleport) {
	for _, handler := range hm.handlers {
		handler.Load().HandleTeleport(e)
	}
}

func (hm *HandlerManager) HandleToggleSneak(e EventToggleSneak) {
	for _, handler := range hm.handlers {
		handler.Load().HandleToggleSneak(e)
	}
}

func (hm *HandlerManager) HandleToggleSprint(e EventToggleSprint) {
	for _, handler := range hm.handlers {
		handler.Load().HandleToggleSprint(e)
	}
}

func (hm *HandlerManager) HandleTransfer(e EventTransfer) {
	for _, handler := range hm.handlers {
		handler.Load().HandleTransfer(e)
	}
}

type NopHandler struct{}

func (NopHandler) HandleAttackEntity(EventAttackEntity)         {}
func (NopHandler) HandleBlockBreak(EventBlockBreak)             {}
func (NopHandler) HandleBlockPick(EventBlockPick)               {}
func (NopHandler) HandleBlockPlace(EventBlockPlace)             {}
func (NopHandler) HandleChangeWorld(EventChangeWorld)           {}
func (NopHandler) HandleChat(EventChat)                         {}
func (NopHandler) HandleCommandExecution(EventCommandExecution) {}
func (NopHandler) HandleDeath(EventDeath)                       {}
func (NopHandler) HandleExperienceGain(EventExperienceGain)     {}
func (NopHandler) HandleFoodLoss(EventFoodLoss)                 {}
func (NopHandler) HandleHeal(EventHeal)                         {}
func (NopHandler) HandleHurt(EventHurt)                         {}
func (NopHandler) HandleItemConsume(EventItemConsume)           {}
func (NopHandler) HandleItemDamage(EventItemDamage)             {}
func (NopHandler) HandleItemDrop(EventItemDrop)                 {}
func (NopHandler) HandleItemPickup(EventItemPickup)             {}
func (NopHandler) HandleItemUse(EventItemUse)                   {}
func (NopHandler) HandleItemUseOnBlock(EventItemUseOnBlock)     {}
func (NopHandler) HandleItemUseOnEntity(EventItemUseOnEntity)   {}
func (NopHandler) HandleJump(EventJump)                         {}
func (NopHandler) HandleMove(EventMove)                         {}
func (NopHandler) HandlePunchAir(EventPunchAir)                 {}
func (NopHandler) HandleQuit(EventQuit)                         {}
func (NopHandler) HandleRespawn(EventRespawn)                   {}
func (NopHandler) HandleSignEdit(EventSignEdit)                 {}
func (NopHandler) HandleSkinChange(EventSkinChange)             {}
func (NopHandler) HandleStartBreak(EventStartBreak)             {}
func (NopHandler) HandleTeleport(EventTeleport)                 {}
func (NopHandler) HandleToggleSneak(EventToggleSneak)           {}
func (NopHandler) HandleToggleSprint(EventToggleSprint)         {}
func (NopHandler) HandleTransfer(EventTransfer)                 {}

func findKey[T any](m map[int]T) int {
	key := 0

	for {
		if _, ok := m[key]; !ok {
			break
		}

		key++
	}

	return key
}

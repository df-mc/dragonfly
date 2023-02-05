package player

import (
	"sync"
)

// HandlerManager manages a player's handlers.
type HandlerManager struct {
	sync.Mutex
	handlers []Handler
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

func (hm *HandlerManager) HandleAttackEntity(e EventAttackEntity) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleAttackEntity(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleBlockBreak(e EventBlockBreak) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleBlockBreak(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleBlockPick(e EventBlockPick) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleBlockPick(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleBlockPlace(e EventBlockPlace) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleBlockPlace(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleChangeWorld(e EventChangeWorld) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleChangeWorld(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleChat(e EventChat) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleChat(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleCommandExecution(e EventCommandExecution) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleCommandExecution(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleDeath(e EventDeath) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleDeath(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleExperienceGain(e EventExperienceGain) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleExperienceGain(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleFoodLoss(e EventFoodLoss) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleFoodLoss(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleHeal(e EventHeal) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleHeal(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleHurt(e EventHurt) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleHurt(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleItemConsume(e EventItemConsume) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleItemConsume(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleItemDamage(e EventItemDamage) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleItemDamage(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleItemDrop(e EventItemDrop) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleItemDrop(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleItemPickup(e EventItemPickup) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleItemPickup(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleItemUse(e EventItemUse) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleItemUse(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleItemUseOnBlock(e EventItemUseOnBlock) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleItemUseOnBlock(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleItemUseOnEntity(e EventItemUseOnEntity) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleItemUseOnEntity(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleJump(e EventJump) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleJump(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleMove(e EventMove) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleMove(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandlePunchAir(e EventPunchAir) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandlePunchAir(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleQuit(e EventQuit) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleQuit(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleRespawn(e EventRespawn) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleRespawn(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleSignEdit(e EventSignEdit) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleSignEdit(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleSkinChange(e EventSkinChange) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleSkinChange(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleStartBreak(e EventStartBreak) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleStartBreak(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleTeleport(e EventTeleport) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleTeleport(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleToggleSneak(e EventToggleSneak) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleToggleSneak(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleToggleSprint(e EventToggleSprint) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleToggleSprint(e)
	}

	hm.Unlock()
}

func (hm *HandlerManager) HandleTransfer(e EventTransfer) {
	hm.Lock()

	for _, h := range hm.handlers {
		h.HandleTransfer(e)
	}

	hm.Unlock()
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

package player

import (
	"github.com/df-mc/atomic"
	"sync"
)

// HandlerManager manages a player's handlers.
type HandlerManager struct {
	sync.Mutex
	handlers []*atomic.Value[Handler]
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

func (hm *HandlerManager) HandleAttackEntity(e EventAttackEntity) {
	for _, h := range hm.handlers {
		h.Load().HandleAttackEntity(e)
	}
}

func (hm *HandlerManager) HandleBlockBreak(e EventBlockBreak) {
	for _, h := range hm.handlers {
		h.Load().HandleBlockBreak(e)
	}
}

func (hm *HandlerManager) HandleBlockPick(e EventBlockPick) {
	for _, h := range hm.handlers {
		h.Load().HandleBlockPick(e)
	}
}

func (hm *HandlerManager) HandleBlockPlace(e EventBlockPlace) {
	for _, h := range hm.handlers {
		h.Load().HandleBlockPlace(e)
	}
}

func (hm *HandlerManager) HandleChangeWorld(e EventChangeWorld) {
	for _, h := range hm.handlers {
		h.Load().HandleChangeWorld(e)
	}
}

func (hm *HandlerManager) HandleChat(e EventChat) {
	for _, h := range hm.handlers {
		h.Load().HandleChat(e)
	}
}

func (hm *HandlerManager) HandleCommandExecution(e EventCommandExecution) {
	for _, h := range hm.handlers {
		h.Load().HandleCommandExecution(e)
	}
}

func (hm *HandlerManager) HandleDeath(e EventDeath) {
	for _, h := range hm.handlers {
		h.Load().HandleDeath(e)
	}
}

func (hm *HandlerManager) HandleExperienceGain(e EventExperienceGain) {
	for _, h := range hm.handlers {
		h.Load().HandleExperienceGain(e)
	}
}

func (hm *HandlerManager) HandleFoodLoss(e EventFoodLoss) {
	for _, h := range hm.handlers {
		h.Load().HandleFoodLoss(e)
	}
}

func (hm *HandlerManager) HandleHeal(e EventHeal) {
	for _, h := range hm.handlers {
		h.Load().HandleHeal(e)
	}
}

func (hm *HandlerManager) HandleHurt(e EventHurt) {
	for _, h := range hm.handlers {
		h.Load().HandleHurt(e)
	}
}

func (hm *HandlerManager) HandleItemConsume(e EventItemConsume) {
	for _, h := range hm.handlers {
		h.Load().HandleItemConsume(e)
	}
}

func (hm *HandlerManager) HandleItemDamage(e EventItemDamage) {
	for _, h := range hm.handlers {
		h.Load().HandleItemDamage(e)
	}
}

func (hm *HandlerManager) HandleItemDrop(e EventItemDrop) {
	for _, h := range hm.handlers {
		h.Load().HandleItemDrop(e)
	}
}

func (hm *HandlerManager) HandleItemPickup(e EventItemPickup) {
	for _, h := range hm.handlers {
		h.Load().HandleItemPickup(e)
	}
}

func (hm *HandlerManager) HandleItemUse(e EventItemUse) {
	for _, h := range hm.handlers {
		h.Load().HandleItemUse(e)
	}
}

func (hm *HandlerManager) HandleItemUseOnBlock(e EventItemUseOnBlock) {
	for _, h := range hm.handlers {
		h.Load().HandleItemUseOnBlock(e)
	}
}

func (hm *HandlerManager) HandleItemUseOnEntity(e EventItemUseOnEntity) {
	for _, h := range hm.handlers {
		h.Load().HandleItemUseOnEntity(e)
	}
}

func (hm *HandlerManager) HandleJump(e EventJump) {
	for _, h := range hm.handlers {
		h.Load().HandleJump(e)
	}
}

func (hm *HandlerManager) HandleMove(e EventMove) {
	for _, h := range hm.handlers {
		h.Load().HandleMove(e)
	}
}

func (hm *HandlerManager) HandlePunchAir(e EventPunchAir) {
	for _, h := range hm.handlers {
		h.Load().HandlePunchAir(e)
	}
}

func (hm *HandlerManager) HandleQuit(e EventQuit) {
	for _, h := range hm.handlers {
		h.Load().HandleQuit(e)
	}
}

func (hm *HandlerManager) HandleRespawn(e EventRespawn) {
	for _, h := range hm.handlers {
		h.Load().HandleRespawn(e)
	}
}

func (hm *HandlerManager) HandleSignEdit(e EventSignEdit) {
	for _, h := range hm.handlers {
		h.Load().HandleSignEdit(e)
	}
}

func (hm *HandlerManager) HandleSkinChange(e EventSkinChange) {
	for _, h := range hm.handlers {
		h.Load().HandleSkinChange(e)
	}
}

func (hm *HandlerManager) HandleStartBreak(e EventStartBreak) {
	for _, h := range hm.handlers {
		h.Load().HandleStartBreak(e)
	}
}

func (hm *HandlerManager) HandleTeleport(e EventTeleport) {
	for _, h := range hm.handlers {
		h.Load().HandleTeleport(e)
	}
}

func (hm *HandlerManager) HandleToggleSneak(e EventToggleSneak) {
	for _, h := range hm.handlers {
		h.Load().HandleToggleSneak(e)
	}
}

func (hm *HandlerManager) HandleToggleSprint(e EventToggleSprint) {
	for _, h := range hm.handlers {
		h.Load().HandleToggleSprint(e)
	}
}

func (hm *HandlerManager) HandleTransfer(e EventTransfer) {
	for _, h := range hm.handlers {
		h.Load().HandleTransfer(e)
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

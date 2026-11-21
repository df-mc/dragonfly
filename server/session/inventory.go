package session

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
)

type heldItemStateUpdater interface {
	UpdateHeldItemState()
}

// shieldUseState records a shield use already handled through PlayerAuthInput so a following inventory transaction
// does not run the item-use handler twice.
type shieldUseState struct {
	pending       bool
	main, offHand item.Stack
}

// markShieldUsePending records the currently held items for a handled shield use.
func (s *Session) markShieldUsePending(c Controllable) {
	s.shieldUse.main, s.shieldUse.offHand = c.HeldItems()
	s.shieldUse.pending = true
}

// consumeShieldUsePending reports whether the next item use matches a shield use already handled through auth input.
func (s *Session) consumeShieldUsePending(c Controllable) bool {
	state := s.shieldUse
	s.shieldUse = shieldUseState{}
	if !state.pending {
		return false
	}
	main, offHand := c.HeldItems()
	return main.Equal(state.main) && offHand.Equal(state.offHand)
}

// clearShieldUsePending discards a previously handled shield use.
func (s *Session) clearShieldUsePending() {
	s.shieldUse = shieldUseState{}
}

// invSlot maps a protocol slot to the inventory's server-side slot: the off hand always uses slot 0.
func (s *Session) invSlot(inv *inventory.Inventory, slot int) int {
	if inv == s.offHand {
		return 0
	}
	return slot
}

// heldItemSlot reports whether the inventory slot is currently held.
func (s *Session) heldItemSlot(inv *inventory.Inventory, slot int) bool {
	return inv == s.offHand || inv == s.inv && s.heldSlot != nil && slot == int(*s.heldSlot)
}

// updateHeldItemState refreshes the controlled entity after held items change.
func (s *Session) updateHeldItemState(tx *world.Tx) {
	if s.ent == nil {
		return
	}
	e, ok := s.ent.Entity(tx)
	if !ok {
		return
	}
	if updater, ok := e.(heldItemStateUpdater); ok {
		updater.UpdateHeldItemState()
	}
}

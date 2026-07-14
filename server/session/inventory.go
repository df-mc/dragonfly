package session

import (
	"github.com/df-mc/dragonfly/server/item/inventory"
	"github.com/df-mc/dragonfly/server/world"
)

type heldItemStateUpdater interface {
	UpdateHeldItemState()
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

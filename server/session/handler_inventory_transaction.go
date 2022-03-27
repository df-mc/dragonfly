package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// InventoryTransactionHandler handles the InventoryTransaction packet.
type InventoryTransactionHandler struct{}

// Handle ...
func (h *InventoryTransactionHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.InventoryTransaction)

	switch data := pk.TransactionData.(type) {
	case *protocol.NormalTransactionData:
		h.resendInventories(s)
		// Always resend inventories with normal transactions. Most of the time we do not use these
		// transactions, so we're best off making sure the client and server stay in sync.
		if err := h.handleNormalTransaction(pk, s); err != nil {
			s.log.Debugf("failed processing packet from %v (%v): InventoryTransaction: failed verifying actions in Normal transaction: %v\n", s.conn.RemoteAddr(), s.c.Name(), err)
			return nil
		}
		return nil
	case *protocol.MismatchTransactionData:
		// Just resend the inventory and don't do anything.
		h.resendInventories(s)
		return nil
	case *protocol.UseItemOnEntityTransactionData:
		if err := s.UpdateHeldSlot(int(data.HotBarSlot), stackToItem(data.HeldItem.Stack)); err != nil {
			return err
		}
		return h.handleUseItemOnEntityTransaction(data, s)
	case *protocol.UseItemTransactionData:
		if err := s.UpdateHeldSlot(int(data.HotBarSlot), stackToItem(data.HeldItem.Stack)); err != nil {
			return err
		}
		return h.handleUseItemTransaction(data, s)
	case *protocol.ReleaseItemTransactionData:
		if err := s.UpdateHeldSlot(int(data.HotBarSlot), stackToItem(data.HeldItem.Stack)); err != nil {
			return err
		}
		return h.handleReleaseItemTransaction(s)
	}
	return fmt.Errorf("unhandled inventory transaction type %T", pk.TransactionData)
}

// resendInventories resends all inventories of the player.
func (h *InventoryTransactionHandler) resendInventories(s *Session) {
	s.sendInv(s.inv, protocol.WindowIDInventory)
	s.sendInv(s.ui, protocol.WindowIDUI)
	s.sendInv(s.offHand, protocol.WindowIDOffHand)
	s.sendInv(s.armour.Inventory(), protocol.WindowIDArmour)
}

// handleNormalTransaction ...
func (h *InventoryTransactionHandler) handleNormalTransaction(pk *packet.InventoryTransaction, s *Session) error {
	for _, action := range pk.Actions {
		switch action.SourceType {
		case protocol.InventoryActionSourceWorld:
			// Item dropping using Q in the hotbar still uses the old inventory transaction system, so we need
			// to account for that.
			if action.OldItem.Stack.Count != 0 || action.OldItem.Stack.NetworkID != 0 || action.OldItem.Stack.MetadataValue != 0 {
				return fmt.Errorf("unexpected non-zero old item in transaction action: %#v", action.OldItem)
			}
			thrown := stackToItem(action.NewItem.Stack)
			held, off := s.c.HeldItems()
			if !thrown.Comparable(held) {
				return fmt.Errorf("different item thrown than held in slot: %#v was thrown but held %#v", thrown, held)
			}
			if thrown.Count() > held.Count() {
				return fmt.Errorf("tried to throw %v items, but held only %v in slot", thrown.Count(), held.Count())
			}

			if err := call(event.C(), int(s.heldSlot.Load()), held.Grow(thrown.Count()-held.Count()), s.inv.Handler().HandleDrop); err != nil {
				return err
			}

			// Explicitly don't re-use the thrown variable. This item was supplied by the user, and if some
			// logic in the Comparable() method was flawed, users would be able to cheat with item properties.
			// Only grow or shrink the held item to prevent any such issues.
			n := s.c.Drop(held.Grow(thrown.Count() - held.Count()))
			s.c.SetHeldItems(held.Grow(-n), off)
		default:
			// Ignore inventory actions we don't explicitly handle.
		}
	}
	return nil
}

// handleUseItemOnEntityTransaction ...
func (h *InventoryTransactionHandler) handleUseItemOnEntityTransaction(data *protocol.UseItemOnEntityTransactionData, s *Session) error {
	s.swingingArm.Store(true)
	defer s.swingingArm.Store(false)

	e, ok := s.entityFromRuntimeID(data.TargetEntityRuntimeID)
	if !ok {
		// In some cases, for example when a falling block entity solidifies, latency may allow attacking an entity that
		// no longer exists server side. This is expected, so we shouldn't kick the player.
		s.log.Debugf("invalid entity interaction: no entity found with runtime ID %v", data.TargetEntityRuntimeID)
		return nil
	}
	if data.TargetEntityRuntimeID == selfEntityRuntimeID {
		return fmt.Errorf("invalid entity interaction: players cannot interact with themselves")
	}
	switch data.ActionType {
	case protocol.UseItemOnEntityActionInteract:
		s.c.UseItemOnEntity(e)
	case protocol.UseItemOnEntityActionAttack:
		s.c.AttackEntity(e)
	default:
		return fmt.Errorf("unhandled UseItemOnEntity ActionType %v", data.ActionType)
	}
	return nil
}

// handleUseItemTransaction ...
func (h *InventoryTransactionHandler) handleUseItemTransaction(data *protocol.UseItemTransactionData, s *Session) error {
	pos := cube.Pos{int(data.BlockPosition[0]), int(data.BlockPosition[1]), int(data.BlockPosition[2])}
	s.swingingArm.Store(true)
	defer s.swingingArm.Store(false)

	// We reset the inventory so that we can send the held item update without the client already
	// having done that client-side.
	// Because of the new inventory system, the client will expect a transaction confirmation, but instead of doing that
	// it's much easier to just resend the inventory.
	h.resendInventories(s)

	switch data.ActionType {
	case protocol.UseItemActionBreakBlock:
		s.c.BreakBlock(pos)
	case protocol.UseItemActionClickBlock:
		s.c.UseItemOnBlock(pos, cube.Face(data.BlockFace), vec32To64(data.ClickedPosition))
	case protocol.UseItemActionClickAir:
		s.c.UseItem()
	default:
		return fmt.Errorf("unhandled UseItem ActionType %v", data.ActionType)
	}
	return nil
}

// handleReleaseItemTransaction ...
func (h *InventoryTransactionHandler) handleReleaseItemTransaction(s *Session) error {
	s.c.ReleaseItem()
	return nil
}

package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/dragonfly/world"
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
		// Always resend inventories with normal transactions. Most of the time we do not use these
		// transactions so we're best off making sure the client and server stay in sync.
		h.resendInventories(s)
		if err := h.handleNormalTransaction(pk, s); err != nil {
			s.log.Debugf("failed processing packet from %v (%v): InventoryTransaction: failed verifying actions in Normal transaction: %w\n", s.conn.RemoteAddr(), s.c.Name(), err)
			return nil
		}
		return nil
	case *protocol.UseItemOnEntityTransactionData:
		return h.handleUseItemOnEntityTransaction(data, s)
	case *protocol.UseItemTransactionData:
		return h.handleUseItemTransaction(data, s)
	}
	return fmt.Errorf("unhandled inventory transaction type %T", pk.TransactionData)
}

// resendInventories resends all inventories of the player.
func (h *InventoryTransactionHandler) resendInventories(s *Session) {
	s.sendInv(s.inv, protocol.WindowIDInventory)
	s.sendInv(s.ui, protocol.WindowIDUI)
	s.sendInv(s.offHand, protocol.WindowIDOffHand)
	s.sendInv(s.armour.Inv(), protocol.WindowIDArmour)
}

// handleNormalTransaction ...
func (h *InventoryTransactionHandler) handleNormalTransaction(pk *packet.InventoryTransaction, s *Session) error {
	for _, action := range pk.Actions {
		switch action.SourceType {
		case protocol.InventoryActionSourceWorld:
			if action.OldItem.Count != 0 || action.OldItem.NetworkID != 0 || action.OldItem.MetadataValue != 0 {
				return fmt.Errorf("unexpected non-zero old item in transaction action: %#v", action.OldItem)
			}
			newItem := stackToItem(action.NewItem)
			actual, offHand := s.c.HeldItems()
			if !newItem.Comparable(actual) {
				return fmt.Errorf("different item thrown than held in hand: %#v was thrown but held %#v", newItem, actual)
			}
			if newItem.Count() > actual.Count() {
				return fmt.Errorf("tried to throw %v items, but held only %v", newItem.Count(), actual.Count())
			}
			// Explicitly don't re-use the newItem variable. This item was supplied by the user, and if some
			// logic in the Comparable() method was flawed, users would be able to cheat with item properties.
			s.c.Drop(actual.Grow(newItem.Count() - actual.Count()))
			s.c.SetHeldItems(actual.Grow(-newItem.Count()), offHand)
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
		return fmt.Errorf("invalid entity interaction: no entity found with runtime ID %v", data.TargetEntityRuntimeID)
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
	pos := world.BlockPos{int(data.BlockPosition[0]), int(data.BlockPosition[1]), int(data.BlockPosition[2])}
	s.swingingArm.Store(true)
	defer s.swingingArm.Store(false)

	switch data.ActionType {
	case protocol.UseItemActionBreakBlock:
		s.c.BreakBlock(pos)
	case protocol.UseItemActionClickBlock:
		// We reset the inventory so that we can send the held item update without the client already
		// having done that client-side.
		s.sendInv(s.inv, protocol.WindowIDInventory)
		s.c.UseItemOnBlock(pos, world.Face(data.BlockFace), vec32To64(data.ClickedPosition))
	case protocol.UseItemActionClickAir:
		s.c.UseItem()
	default:
		return fmt.Errorf("unhandled UseItem ActionType %v", data.ActionType)
	}
	return nil
}

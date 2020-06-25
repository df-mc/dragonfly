package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"sync/atomic"
)

// InventoryTransactionHandler handles the InventoryTransaction packet.
type InventoryTransactionHandler struct{}

// Handle ...
func (h *InventoryTransactionHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.InventoryTransaction)

	switch data := pk.TransactionData.(type) {
	case *protocol.NormalTransactionData:
		h.resendInventories(s)
		s.log.Debugf("failed processing packet from %v (%v): InventoryTransaction: unhandled normal transaction %#v\n", s.conn.RemoteAddr(), s.c.Name(), data)
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

// handleUseItemOnEntityTransaction
func (h *InventoryTransactionHandler) handleUseItemOnEntityTransaction(data *protocol.UseItemOnEntityTransactionData, s *Session) error {
	atomic.StoreUint32(s.swingingArm, 1)
	defer atomic.StoreUint32(s.swingingArm, 0)

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

// handleUseItemTransaction
func (h *InventoryTransactionHandler) handleUseItemTransaction(data *protocol.UseItemTransactionData, s *Session) error {
	pos := world.BlockPos{int(data.BlockPosition[0]), int(data.BlockPosition[1]), int(data.BlockPosition[2])}
	atomic.StoreUint32(s.swingingArm, 1)
	defer atomic.StoreUint32(s.swingingArm, 0)

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

package session

import (
	"fmt"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/inventory"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world/gamemode"
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
		return h.handleNormalTransaction(pk, s)
	case *protocol.UseItemOnEntityTransactionData:
		return h.handleUseItemOnEntityTransaction(data, s)
	case *protocol.UseItemTransactionData:
		return h.handleUseItemTransaction(data, s)
	}
	return fmt.Errorf("unhandled inventory transaction type %T", pk.TransactionData)
}

// handleNormalTransaction handles a normal transaction (moving items between inventories, etc).
func (h *InventoryTransactionHandler) handleNormalTransaction(pk *packet.InventoryTransaction, s *Session) error {
	if len(pk.Actions) == 0 {
		return nil
	}
	if err := verifyTransaction(pk.Actions, s); err != nil {
		s.sendInv(s.inv, protocol.WindowIDInventory)
		s.sendInv(s.ui, protocol.WindowIDUI)
		s.sendInv(s.offHand, protocol.WindowIDOffHand)
		s.log.Debugf("%v: %v", s.c.Name(), err)
		return nil
	}
	atomic.StoreUint32(s.inTransaction, 1)
	executeTransaction(pk.Actions, s)
	atomic.StoreUint32(s.inTransaction, 0)
	return nil
}

// handleUseItemOnEntityTransaction
func (h *InventoryTransactionHandler) handleUseItemOnEntityTransaction(data *protocol.UseItemOnEntityTransactionData, s *Session) error {
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
	switch data.ActionType {
	case protocol.UseItemActionBreakBlock:
		s.c.BreakBlock(pos)
	case protocol.UseItemActionClickBlock:
		// We reset the inventory so that we can send the held item update without the client already
		// having done that client-side.
		s.sendInv(s.inv, protocol.WindowIDInventory)
		s.c.UseItemOnBlock(pos, world.Face(data.BlockFace), data.ClickedPosition)
	case protocol.UseItemActionClickAir:
		s.c.UseItem()
	default:
		return fmt.Errorf("unhandled UseItem ActionType %v", data.ActionType)
	}
	return nil
}

// verifyTransaction verifies a transaction composed of the actions passed. The method makes sure the old
// items are precisely equal to the new items: No new items must be added or removed.
// verifyTransaction also checks if all window IDs sent match some inventory, and if the old items match the
// items found in that inventory.
func verifyTransaction(a []protocol.InventoryAction, s *Session) error {
	// Allocate a big inventory and add all new items to it.
	temp := inventory.New(128, nil)
	actions := make([]protocol.InventoryAction, 0, len(a))
	for _, action := range a {
		if action.OldItem.Count == 0 && action.NewItem.Count == 0 {
			continue
		}
		actions = append(actions, action)
	}
	for _, action := range actions {
		inv, ok := s.invByID(action.WindowID)
		if !ok {
			// The inventory with that window ID did not exist.
			return fmt.Errorf("unknown inventory ID %v", action.WindowID)
		}
		actualOld, err := inv.Item(int(action.InventorySlot))
		if err != nil {
			// The slot passed actually exceeds the inventory size, meaning we can't actually get an item at
			// that slot.
			return fmt.Errorf("slot %v out of range for inventory %v", action.InventorySlot, action.WindowID)
		}
		old := stackToItem(action.OldItem)
		if !actualOld.Comparable(old) || actualOld.Count() != old.Count() {
			if _, creative := s.c.GameMode().(gamemode.Creative); !creative || action.SourceType != protocol.InventoryActionSourceCreative {
				// Either the type or the count of the old item as registered by the server and the client are
				// mismatched.
				return fmt.Errorf("slot %v holds a different item than the client expects: %v is actually %v", action.InventorySlot, old, actualOld)
			}
		}
		if _, err := temp.AddItem(old); err != nil {
			return fmt.Errorf("inventory was full: %w", err)
		}
	}
	for _, action := range actions {
		newItem := stackToItem(action.NewItem)
		if err := temp.RemoveItem(newItem); err != nil {
			return fmt.Errorf("item %v removed was not present in old items: %w", newItem, err)
		}
	}
	// Now that we made sure every new item was also present in the old items, we must also check if every old
	// item is present as a new item. We can do that by simply checking if the inventory has any items left in
	// it.
	if !temp.Empty() {
		return fmt.Errorf("new items and old items must be balanced")
	}
	return nil
}

// executeTransaction executes all actions of a transaction passed. It assumes the actions are all valid,
// which would otherwise be checked by calling verifyTransaction.
func executeTransaction(actions []protocol.InventoryAction, s *Session) {
	for _, action := range actions {
		if action.SourceType == protocol.InventoryActionSourceCreative {
			continue
		}
		// The window IDs are already checked when using verifyTransaction, so we don't need to check again.
		inv, _ := s.invByID(action.WindowID)
		_ = inv.SetItem(int(action.InventorySlot), stackToItem(action.NewItem))
	}
}

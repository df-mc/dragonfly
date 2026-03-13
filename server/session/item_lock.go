package session

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
)

func isPlayerInventoryContainer(id int32) bool {
	switch id {
	case protocol.ContainerHotBar, protocol.ContainerInventory, protocol.ContainerCombinedHotBarAndInventory,
		protocol.ContainerOffhand, protocol.ContainerArmor, protocol.ContainerCursor:
		return true
	default:
		return false
	}
}

func (s *Session) sameInventorySlot(a, b protocol.StackRequestSlotInfo, tx *world.Tx) bool {
	invA, okA := s.invByID(int32(a.Container.ContainerID), tx)
	invB, okB := s.invByID(int32(b.Container.ContainerID), tx)
	if !okA || !okB || invA != invB {
		return false
	}
	if invA == s.offHand {
		return true
	}
	return a.Slot == b.Slot
}

func ensureUnlockedForInventoryMove(stack item.Stack, from, to protocol.StackRequestSlotInfo, s *Session, tx *world.Tx) error {
	if stack.Empty() || !isPlayerInventoryContainer(int32(from.Container.ContainerID)) {
		return nil
	}
	switch stack.LockMode() {
	case item.LockInInventory:
		if !isPlayerInventoryContainer(int32(to.Container.ContainerID)) {
			return fmt.Errorf("item is locked in inventory")
		}
	case item.LockInSlot:
		if !s.sameInventorySlot(from, to, tx) {
			return fmt.Errorf("item is locked in slot")
		}
	}
	return nil
}

func ensureUnlockedForInventoryRemoval(stack item.Stack, from protocol.StackRequestSlotInfo) error {
	if stack.Empty() || !isPlayerInventoryContainer(int32(from.Container.ContainerID)) || !stack.Locked() {
		return nil
	}
	switch stack.LockMode() {
	case item.LockInInventory:
		return fmt.Errorf("item is locked in inventory")
	case item.LockInSlot:
		return fmt.Errorf("item is locked in slot")
	default:
		return nil
	}
}

func ensureUnlockedForCrafting(stack item.Stack) error {
	if !stack.Locked() {
		return nil
	}
	return fmt.Errorf("item is locked and cannot be used as a crafting ingredient")
}

func ensureUnlockedForAnvil(stack item.Stack) error {
	if !stack.Locked() {
		return nil
	}
	return fmt.Errorf("item is locked and cannot be renamed or combined in an anvil")
}

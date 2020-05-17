package session

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"sync/atomic"
)

// MobEquipmentHandler handles the MobEquipment packet.
type MobEquipmentHandler struct{}

// Handle ...
func (*MobEquipmentHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.MobEquipment)

	if pk.EntityRuntimeID != selfEntityRuntimeID {
		return ErrSelfRuntimeID
	}
	// The slot that the player might have selected must be within the hotbar: The held item cannot be in a
	// different place in the inventory.
	if pk.InventorySlot > 8 {
		return fmt.Errorf("slot exceeds hotbar range 0-8: slot is %v", pk.InventorySlot)
	}
	if pk.WindowID != protocol.WindowIDInventory {
		return fmt.Errorf("only main inventory should be involved, got window ID %v", pk.WindowID)
	}
	clientSideItem := stackToItem(pk.NewItem)
	actual, _ := s.inv.Item(int(pk.InventorySlot))

	// The item the client claims to have must be identical to the one we have registered server-side.
	if !clientSideItem.Comparable(actual) {
		// Only ever debug these as they are frequent and expected to happen whenever client and server get
		// out of sync.
		s.log.Debugf("failed processing packet from %v (%v): *packet.MobEquipment: client-side item must be identical to server-side item, but got different types: client: %v vs server: %v", s.conn.RemoteAddr(), s.c.Name(), clientSideItem, actual)
	}
	if clientSideItem.Count() != actual.Count() {
		// Only ever debug these as they are frequent and expected to happen whenever client and server get
		// out of sync.
		s.log.Debugf("failed processing packet from %v (%v): *packet.MobEquipment: client-side item must be identical to server-side item, but got different counts: client: %v vs server: %v", s.conn.RemoteAddr(), s.c.Name(), clientSideItem.Count(), actual.Count())
	}

	// We first change the held slot.
	atomic.StoreUint32(s.heldSlot, uint32(pk.InventorySlot))

	for _, viewer := range s.c.World().Viewers(s.c.Position()) {
		viewer.ViewEntityItems(s.c)
	}
	return nil
}

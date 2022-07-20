package session

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// MobEquipmentHandler handles the MobEquipment packet.
type MobEquipmentHandler struct{}

// Handle ...
func (*MobEquipmentHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.MobEquipment)

	if pk.EntityRuntimeID != selfEntityRuntimeID {
		return errSelfRuntimeID
	}
	switch pk.WindowID {
	case protocol.WindowIDOffHand:
		// This window ID is expected, but we don't handle it.
		return nil
	case protocol.WindowIDInventory:
		// The item the client claims to have must be identical to the one we have registered server-side.
		actual, _ := s.inv.Item(int(pk.InventorySlot))
		clientSide := stackToItem(pk.NewItem.Stack)
		if !actual.Equal(clientSide) {
			// Only ever debug these as they are frequent and expected to happen whenever client and server get
			// out of sync.
			s.log.Debugf("failed processing packet from %v (%v): *packet.MobEquipment: client-side item must be identical to server-side item, but got differences: client: %v vs server: %v", s.conn.RemoteAddr(), s.c.Name(), clientSide, actual)
		}
		return s.c.SetHeldSlot(int(pk.InventorySlot))
	default:
		return fmt.Errorf("only main inventory should be involved in slot chnage, got window ID %v", pk.WindowID)
	}
}

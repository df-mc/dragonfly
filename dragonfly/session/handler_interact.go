package session

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// InteractHandler handles the Interact packet.
type InteractHandler struct{}

// Handle ...
func (h *InteractHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.Interact)

	switch pk.ActionType {
	case packet.InteractActionMouseOverEntity:
		// We don't need this action.
	case packet.InteractActionOpenInventory:
		s.writePacket(&packet.ContainerOpen{
			WindowID:      0,
			ContainerType: 0xff,
		})
	default:
		return fmt.Errorf("unexpected interact packet action %v", pk.ActionType)
	}
	return nil
}

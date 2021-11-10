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
		if s.invOpened {
			// When there is latency, this might end up being sent multiple times. If we send a ContainerOpen
			// multiple times, the client crashes.
			return nil
		}
		s.invOpened = true
		s.writePacket(&packet.ContainerOpen{
			WindowID:      0,
			ContainerType: 0xff,
		})
		break
	case packet.InteractActionLeaveVehicle:
		if s.c.Riding() == nil {
			return nil
		}
		s.c.DismountEntity(s.c.Riding())
		break
	default:
		return fmt.Errorf("unexpected interact packet action %v", pk.ActionType)
	}
	return nil
}

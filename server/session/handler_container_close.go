package session

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ContainerCloseHandler handles the ContainerClose packet.
type ContainerCloseHandler struct{}

// Handle ...
func (h *ContainerCloseHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.ContainerClose)

	s.EmptyUIInventory()
	switch pk.WindowID {
	case 0:
		// Closing of the normal inventory.
		s.writePacket(&packet.ContainerClose{WindowID: 0})
		s.invOpened = false
	case byte(s.openedWindowID.Load()):
		s.closeCurrentContainer()
	case 0xff:
		// TODO: Handle closing the crafting grid.
	default:
		return fmt.Errorf("unexpected close request for unopened container %v", pk.WindowID)
	}
	return nil
}

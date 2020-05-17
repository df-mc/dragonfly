package session

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"sync/atomic"
)

// ContainerCloseHandler handles the ContainerClose packet.
type ContainerCloseHandler struct{}

// Handle ...
func (h *ContainerCloseHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.ContainerClose)

	switch pk.WindowID {
	case byte(atomic.LoadUint32(&s.openedWindowID)):
		s.closeCurrentContainer()
	case 255:
		// TODO: Handle closing the crafting grid.
	default:
		return fmt.Errorf("unexpected close request for unopened container %v", pk.WindowID)
	}
	return nil
}

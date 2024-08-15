package session

import (
	"fmt"
	"github.com/df-mc/atomic"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ServerBoundLoadingScreenHandler handles loading screen updates from the clients. It is used to ensure that
// the server knows when the client is loading a screen, and when it is done loading it.
type ServerBoundLoadingScreenHandler struct {
	currentID  atomic.Uint32
	expectedID atomic.Uint32
}

// Handle ...
func (h *ServerBoundLoadingScreenHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.ServerBoundLoadingScreen)
	v, ok := pk.LoadingScreenID.Value()
	if !ok || h.expectedID.Load() == 0 {
		return nil
	} else if v != h.expectedID.Load() {
		return fmt.Errorf("expected loading screen ID %d, got %d", h.expectedID.Load(), v)
	} else if pk.Type == packet.LoadingScreenTypeEnd {
		s.changingDimension.Store(false)
		h.expectedID.Store(0)
	}
	return nil
}

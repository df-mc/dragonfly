package session

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PlayerSkinHandler handles the PlayerSkin packet.
type PlayerSkinHandler struct{}

// Handle ...
func (h *PlayerSkinHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.PlayerSkin)
	s.BroadcastSkinChange(pk)
	return nil
}

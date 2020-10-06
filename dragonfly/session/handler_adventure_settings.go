package session

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// AdventureSettingsHandler handles the AdventureSettings packet.
type AdventureSettingsHandler struct{}

// Handle ...
func (b AdventureSettingsHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.AdventureSettings)
	s.adventureFlags.Store(pk.Flags)
	return nil
}

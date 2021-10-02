package session

import "github.com/sandertv/gophertunnel/minecraft/protocol/packet"

// AdventureSettingsHandler handles the AdventureSettings packet.
type AdventureSettingsHandler struct{}

// Handle ...
func (b AdventureSettingsHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.AdventureSettings)
	if pk.Flags&packet.AdventureFlagFlying != 0 && !s.c.GameMode().AllowsFlying() || pk.Flags&packet.AdventureFlagWorldBuilder != 0 && !s.c.GameMode().AllowsEditing() || pk.Flags&packet.AdventureFlagNoPVP != 0 && !s.c.GameMode().AllowsInteraction() {
		// Resend the adventure settings packet.
		s.sendAdventureSettings(s.c.GameMode())
	}
	return nil
}

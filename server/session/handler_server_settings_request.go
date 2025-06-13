package session

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ServerSettingsRequestHandler ...
type ServerSettingsRequestHandler struct{}

// Handle ...
func (h ServerSettingsRequestHandler) Handle(_ packet.Packet, s *Session, _ *world.Tx, _ Controllable) error {
	if !s.HasServerSettingsForm() {
		return nil
	}
	s.SendServerSettingsForm()
	return nil
}

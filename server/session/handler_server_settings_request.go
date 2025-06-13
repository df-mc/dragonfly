package session

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ServerSettingsRequestHandler handles a server settings request from the client.
// If the session has a server settings form attached, it sends the form to the client.
type ServerSettingsRequestHandler struct{}

// Handle ...
func (h ServerSettingsRequestHandler) Handle(_ packet.Packet, s *Session, _ *world.Tx, _ Controllable) error {
	if !s.HasServerSettingsForm() {
		return nil
	}
	s.SendServerSettingsForm()
	return nil
}

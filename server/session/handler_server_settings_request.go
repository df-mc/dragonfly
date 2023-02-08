package session

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ServerSettingsRequest Is send by the client in order to request a ServerSettingsResponse packet
type ServerSettingsRequest struct{}

// Handle ...
func (e ServerSettingsRequest) Handle(_ packet.Packet, s *Session) error {
	if !s.hasSettingsResponse.Load() {
		return nil
	}
	s.writePacket(&packet.ServerSettingsResponse{
		FormData: s.settingsResponse.Load(),
		FormID:   s.settingsFormID.Load(),
	})
	return nil
}

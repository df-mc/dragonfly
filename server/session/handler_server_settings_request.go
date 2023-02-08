package session

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ServerSettingsRequest Responds with a ServerSettings packet containing a form.
// The form sent back will be shown in the settings menu at the top of the settings.
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

package session

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ServerSettingsRequest Is send by the client
// in order to request a ServerSettingsResponse packet
type ServerSettingsRequest struct{}

// Handle Sends the ServerSettingsResponse packet.
func (e ServerSettingsRequest) Handle(_ packet.Packet, s *Session) error {
	if !s.shouldSendSettingsResponse {
		return nil
	}
	s.writePacket(&packet.ServerSettingsResponse{
		FormData: s.serverSettingsResponse,
		FormID:   s.serverSettingsFormID,
	})
	return nil
}

package session

import (
	"github.com/google/uuid"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// PlayerSkinHandler handles the PlayerSkin packet.
type PlayerSkinHandler struct{}

// Handle ...
func (b PlayerSkinHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.PlayerSkin)
	puuid, err := uuid.Parse(s.conn.IdentityData().Identity)
	if err != nil { // the session has an invalid uuid?
		return err
	}
	s.SetSkin(protocolToSkin(pk.Skin), puuid)
	return nil
}

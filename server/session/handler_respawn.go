package session

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// RespawnHandler handles the Respawn packet.
type RespawnHandler struct{}

// Handle ...
func (*RespawnHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.Respawn)

	if pk.EntityRuntimeID != selfEntityRuntimeID {
		return ErrSelfRuntimeID
	}
	//noinspection GoCommentLeadingSpace
	if pk.State != packet.RespawnStateClientReadyToSpawn {
		//lint:ignore ST1005 Error string is only capitalised because of the field name.
		//noinspection GoErrorStringFormat
		return fmt.Errorf("State must always be %v, but got %v", packet.RespawnStateClientReadyToSpawn, pk.State)
	}
	s.c.Respawn()
	return nil
}

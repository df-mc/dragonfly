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
		return errSelfRuntimeID
	}
	if pk.State != packet.RespawnStateClientReadyToSpawn {
		return fmt.Errorf("respawn state must always be %v, but got %v", packet.RespawnStateClientReadyToSpawn, pk.State)
	}
	s.c.Respawn()
	return nil
}

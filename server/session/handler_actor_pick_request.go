package session

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// ActorPickRequestHandler handles the ActorPickRequest packet.
type ActorPickRequestHandler struct{}

// Handle ...
func (ActorPickRequestHandler) Handle(p packet.Packet, s *Session, tx *world.Tx, c Controllable) error {
	pk := p.(*packet.ActorPickRequest)
	handle, ok := s.entityFromRuntimeID(uint64(pk.EntityUniqueID))
	if !ok {
		return nil
	}
	if e, ok := handle.Entity(tx); ok {
		c.PickEntity(e)
	}
	return nil
}

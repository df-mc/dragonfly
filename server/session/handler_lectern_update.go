package session

import (
	"fmt"
	"github.com/df-mc/dragonfly/server/block"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// LecternUpdateHandler handles the LecternUpdate packet, sent when a player interacts with a lectern.
type LecternUpdateHandler struct{}

// Handle ...
func (LecternUpdateHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.LecternUpdate)
	pos := blockPosFromProtocol(pk.Position)
	if !canReach(s.c, pos.Vec3Middle()) {
		return fmt.Errorf("block at %v is not within reach", pos)
	}
	if _, ok := s.c.World().Block(pos).(block.Lectern); !ok {
		return fmt.Errorf("block at %v is not a lectern", pos)
	}
	return s.c.TurnLecternPage(pos, int(pk.Page))
}

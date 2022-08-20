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
	if pk.DropBook {
		// This is completely redundant, so ignore this packet.
		return nil
	}
	pos := blockPosFromProtocol(pk.Position)
	if !s.c.CanReach(pos.Vec3Middle()) {
		return fmt.Errorf("block at %v is not within reach", pos)
	}
	w := s.c.World()
	lectern, ok := w.Block(pos).(block.Lectern)
	if !ok {
		return fmt.Errorf("block at %v is not a lectern", pos)
	}
	return lectern.TurnPage(pos, w, int(pk.Page))
}

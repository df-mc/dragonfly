package session

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// SubChunkRequestHandler handles sub-chunk requests from the client. The server will respond with a packet containing
// the requested sub-chunks.
type SubChunkRequestHandler struct{}

// Handle ...
func (*SubChunkRequestHandler) Handle(p packet.Packet, s *Session, tx *world.Tx, _ Controllable) error {
	pk := p.(*packet.SubChunkRequest)
	if dimID, _ := world.DimensionID(tx.World().Dimension()); pk.Dimension != int32(dimID) {
		// Outdated sub chunk request from a previous dimension.
		s.writePacket(&packet.SubChunk{
			Dimension:       pk.Dimension,
			Position:        pk.Position,
			CacheEnabled:    s.conn.ClientCacheEnabled(),
			SubChunkEntries: []protocol.SubChunkEntry{},
		})
		return nil
	}
	s.ViewSubChunks(world.SubChunkPos(pk.Position), pk.Offsets, tx)
	return nil
}

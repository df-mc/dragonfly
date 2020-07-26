package session

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// RequestChunkRadiusHandler handles the RequestChunkRadius packet.
type RequestChunkRadiusHandler struct{}

// Handle ...
func (*RequestChunkRadiusHandler) Handle(p packet.Packet, s *Session) error {
	pk := p.(*packet.RequestChunkRadius)

	if pk.ChunkRadius > s.maxChunkRadius {
		pk.ChunkRadius = s.maxChunkRadius
	}
	s.chunkRadius = pk.ChunkRadius

	s.chunkLoader.ChangeRadius(int(pk.ChunkRadius))

	s.writePacket(&packet.ChunkRadiusUpdated{ChunkRadius: s.chunkRadius})
	return nil
}

package session

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// chunkInteractionReady reports if the block position is in a chunk that was
// already sent to this session. Client-driven block actions must pass this check
// before touching world state that may otherwise synchronously load a chunk.
func (s *Session) chunkInteractionReady(pos cube.Pos) bool {
	return s.chunkPosInteractionReady(world.ChunkPos{int32(pos[0] >> 4), int32(pos[2] >> 4)})
}

// positionInteractionReady reports if an entity/player position is in a chunk
// this session can safely interact with.
func (s *Session) positionInteractionReady(pos mgl64.Vec3) bool {
	return s.chunkPosInteractionReady(world.ChunkPos{int32(pos[0]) >> 4, int32(pos[2]) >> 4})
}

func (s *Session) chunkPosInteractionReady(pos world.ChunkPos) bool {
	return s.chunkLoader.Loaded(pos)
}

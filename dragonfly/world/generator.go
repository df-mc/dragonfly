package world

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/block"
	"github.com/dragonfly-tech/dragonfly/dragonfly/world/chunk"
)

// Generator handles the generating of newly created chunks. Worlds have one generator which is used to
// generate chunks when the provider of the world cannot find a chunk at a given chunk position.
type Generator interface {
	// GenerateChunk generates a chunk at a chunk position passed. The generator sets blocks in the chunk that
	// is passed to the method.
	GenerateChunk(pos ChunkPos, chunk *chunk.Chunk)
}

// FlatGenerator is the default generator of World. It generates flat worlds (like those in vanilla) with no
// other decoration.
type FlatGenerator struct{}

var (
	grass, _   = block.RuntimeID(block.Grass{})
	dirt, _    = block.RuntimeID(block.Dirt{})
	bedrock, _ = block.RuntimeID(block.Bedrock{})
)

// GenerateChunk ...
func (FlatGenerator) GenerateChunk(pos ChunkPos, chunk *chunk.Chunk) {
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			chunk.SetRuntimeID(x, 0, z, 0, bedrock)
			chunk.SetRuntimeID(x, 1, z, 0, dirt)
			chunk.SetRuntimeID(x, 2, z, 0, dirt)
			chunk.SetRuntimeID(x, 3, z, 0, grass)
		}
	}
}

package generator

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
)

// Flat is the flat generator of World. It generates flat worlds (like those in vanilla) with no other
// decoration.
// The Layers field may be used to specify the block layers placed.
type Flat struct {
	// Layers is a list of block layers placed by the Flat generator. The layers are ordered in a way where the last
	// element in the slice is placed as the bottom most block of the chunk.
	Layers []world.Block
}

// GenerateChunk ...
func (f Flat) GenerateChunk(_ world.ChunkPos, chunk *chunk.Chunk) {
	// Get a list of block runtime IDs.
	l := int16(len(f.Layers))
	m := make([]uint32, l)
	for i, b := range f.Layers {
		m[i], _ = world.BlockRuntimeID(b)
	}

	min := int16(chunk.Range()[0])
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			for y := int16(0); y < l; y++ {
				chunk.SetBlock(x, min+y, z, 0, m[l-y-1])
			}
		}
	}
}

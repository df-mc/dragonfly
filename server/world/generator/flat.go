package generator

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/biome"
	"github.com/df-mc/dragonfly/server/world/chunk"
)

// Flat is the flat generator of World. It generates flat worlds (like those in vanilla) with no other
// decoration.
// The Layers field may be used to specify the block layers placed.
type Flat struct {
	// Biome is the biome that the generator should use.
	Biome biome.Biome
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

	b := uint32(f.Biome.EncodeBiome())
	min := int16(chunk.Range().Min())
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			for y := int16(0); y < int16(chunk.Range().Max()); y++ {
				if y < l {
					chunk.SetBlock(x, min+y, z, 0, m[l-y-1])
				}
				chunk.SetBiome(x, min+y, z, b)
			}
		}
	}
}

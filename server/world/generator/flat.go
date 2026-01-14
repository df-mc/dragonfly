package generator

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/chunk"
)

// Flat is the flat generator of World. It generates flat worlds (like those in vanilla) with no other
// decoration. It may be constructed by calling NewFlat.
type Flat struct {
	// biome is the encoded biome that the generator should use.
	biome uint32
	// layers is a list of block runtime ID layers placed by the Flat generator. The layers are ordered in a way where
	// the last element in the slice is placed as the bottom-most block of the chunk.
	layers []world.Block
	// n is the amount of layers in the slice above.
	n int16
}

// NewFlat creates a new Flat generator. Chunks generated are completely filled with the world.Biome passed. layers is a
// list of block layers placed by the Flat generator. The layers are ordered in a way where the last element in the
// slice is placed as the bottom-most block of the chunk.
func NewFlat(biome world.Biome, layers []world.Block) Flat {
	f := Flat{
		biome:  uint32(biome.EncodeBiome()),
		layers: layers,
		n:      int16(len(layers)),
	}
	return f
}

// GenerateChunk ...
func (f Flat) GenerateChunk(_ world.ChunkPos, chunk *chunk.Chunk) {
	br := chunk.BlockRegistry.(world.BlockRegistry)
	// Resolve runtime IDs once per chunk generation call. Runtime IDs are registry-specific,
	// so this can't be done in NewFlat.
	layerRIDs := make([]uint32, len(f.layers))
	for i, b := range f.layers {
		layerRIDs[i] = br.BlockRuntimeID(b)
	}

	min, max := int16(chunk.Range().Min()), int16(chunk.Range().Max())
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			for y := int16(0); y <= max; y++ {
				if y < f.n {
					chunk.SetBlock(x, min+y, z, 0, layerRIDs[int(f.n-y-1)])
				}
				chunk.SetBiome(x, min+y, z, f.biome)
			}
		}
	}
}

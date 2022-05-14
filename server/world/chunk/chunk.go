package chunk

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"sync"
)

// Chunk is a segment in the world with a size of 16x16x256 blocks. A chunk contains multiple sub chunks
// and stores other information such as biomes.
// It is not safe to call methods on Chunk simultaneously from multiple goroutines.
type Chunk struct {
	sync.Mutex
	// r holds the (vertical) range of the Chunk. It includes both the minimum and maximum coordinates.
	r cube.Range
	// air is the runtime ID of air.
	air uint32
	// recalculateHeightMap is true if the chunk's height map should be recalculated on the next call to the HeightMap
	// function.
	recalculateHeightMap bool
	// heightMap is the height map of the chunk.
	heightMap HeightMap
	// sub holds all sub chunks part of the chunk. The pointers held by the array are nil if no sub chunk is
	// allocated at the indices.
	sub []*SubChunk
	// biomes is an array of biome IDs. There is one biome ID for every column in the chunk.
	biomes []*PalettedStorage
}

// New initialises a new chunk and returns it, so that it may be used.
func New(air uint32, r cube.Range) *Chunk {
	n := (r.Height() >> 4) + 1
	sub, biomes := make([]*SubChunk, n), make([]*PalettedStorage, n)
	for i := 0; i < n; i++ {
		sub[i] = NewSubChunk(air)
		biomes[i] = emptyStorage(0)
	}
	return &Chunk{
		r:                    r,
		air:                  air,
		sub:                  sub,
		biomes:               biomes,
		recalculateHeightMap: true,
		heightMap:            make(HeightMap, 256),
	}
}

// Range returns the cube.Range of the Chunk as passed to New.
func (chunk *Chunk) Range() cube.Range {
	return chunk.r
}

// Sub returns a list of all sub chunks present in the chunk.
func (chunk *Chunk) Sub() []*SubChunk {
	return chunk.sub
}

// Block returns the runtime ID of the block at a given x, y and z in a chunk at the given layer. If no
// sub chunk exists at the given y, the block is assumed to be air.
func (chunk *Chunk) Block(x uint8, y int16, z uint8, layer uint8) uint32 {
	sub := chunk.SubChunk(y)
	if sub.Empty() || uint8(len(sub.storages)) <= layer {
		return chunk.air
	}
	return sub.storages[layer].At(x, uint8(y), z)
}

// SetBlock sets the runtime ID of a block at a given x, y and z in a chunk at the given layer. If no
// SubChunk exists at the given y, a new SubChunk is created and the block is set.
func (chunk *Chunk) SetBlock(x uint8, y int16, z uint8, layer uint8, block uint32) {
	sub := chunk.sub[chunk.SubIndex(y)]
	if uint8(len(sub.storages)) <= layer && block == chunk.air {
		// Air was set at n layer, but there were less than n layers, so there already was air there.
		// Don't do anything with this, just return.
		return
	}
	sub.Layer(layer).Set(x, uint8(y), z, block)
	chunk.recalculateHeightMap = true
}

// Biome returns the biome ID at a specific column in the chunk.
func (chunk *Chunk) Biome(x uint8, y int16, z uint8) uint32 {
	return chunk.biomes[chunk.SubIndex(y)].At(x, uint8(y), z)
}

// SetBiome sets the biome ID at a specific column in the chunk.
func (chunk *Chunk) SetBiome(x uint8, y int16, z uint8, biome uint32) {
	chunk.biomes[chunk.SubIndex(y)].Set(x, uint8(y), z, biome)
}

// Light returns the light level at a specific position in the chunk.
func (chunk *Chunk) Light(x uint8, y int16, z uint8) uint8 {
	ux, uy, uz, sub := x&0xf, uint8(y&0xf), z&0xf, chunk.SubChunk(y)
	sky := sub.SkyLight(ux, uy, uz)
	if sky == 15 {
		// The skylight was already on the maximum value, so return it without checking block light.
		return sky
	}
	if block := sub.BlockLight(ux, uy, uz); block > sky {
		return block
	}
	return sky
}

// SkyLight returns the skylight level at a specific position in the chunk.
func (chunk *Chunk) SkyLight(x uint8, y int16, z uint8) uint8 {
	return chunk.SubChunk(y).SkyLight(x&15, uint8(y&15), z&15)
}

// HighestLightBlocker iterates from the highest non-empty sub chunk downwards to find the Y value of the
// highest block that completely blocks any light from going through. If none is found, the value returned is
// the minimum height.
func (chunk *Chunk) HighestLightBlocker(x, z uint8) int16 {
	for index := int16(len(chunk.sub) - 1); index >= 0; index-- {
		if sub := chunk.sub[index]; !sub.Empty() {
			for y := 15; y >= 0; y-- {
				if FilteringBlocks[sub.storages[0].At(x, uint8(y), z)] == 15 {
					return int16(y) | chunk.SubY(index)
				}
			}
		}
	}
	return int16(chunk.r[0])
}

// HighestBlock iterates from the highest non-empty sub chunk downwards to find the Y value of the highest
// non-air block at an x and z. If no blocks are present in the column, the minimum height is returned.
func (chunk *Chunk) HighestBlock(x, z uint8) int16 {
	for index := int16(len(chunk.sub) - 1); index >= 0; index-- {
		if sub := chunk.sub[index]; !sub.Empty() {
			for y := 15; y >= 0; y-- {
				if rid := sub.storages[0].At(x, uint8(y), z); rid != chunk.air {
					return int16(y) | chunk.SubY(index)
				}
			}
		}
	}
	return int16(chunk.r[0])
}

// HeightMap returns the height map of the chunk. If the chunk is edited, the height map will be recalculated on the
// next call to this function.
func (chunk *Chunk) HeightMap() HeightMap {
	if chunk.recalculateHeightMap {
		for x := uint8(0); x < 16; x++ {
			for z := uint8(0); z < 16; z++ {
				chunk.heightMap.Set(x, z, chunk.HighestLightBlocker(x, z))
			}
		}
		chunk.recalculateHeightMap = false
	}
	return chunk.heightMap
}

// Compact compacts the chunk as much as possible, getting rid of any sub chunks that are empty, and compacts
// all storages in the sub chunks to occupy as little space as possible.
// Compact should be called right before the chunk is saved in order to optimise the storage space.
func (chunk *Chunk) Compact() {
	for i := range chunk.sub {
		chunk.sub[i].compact()
	}
}

// SubChunk finds the correct SubChunk in the Chunk by a Y value.
func (chunk *Chunk) SubChunk(y int16) *SubChunk {
	return chunk.sub[chunk.SubIndex(y)]
}

// SubIndex returns the sub chunk Y index matching the y value passed.
func (chunk *Chunk) SubIndex(y int16) int16 {
	return (y - int16(chunk.r[0])) >> 4
}

// SubY returns the sub chunk Y value matching the index passed.
func (chunk *Chunk) SubY(index int16) int16 {
	return (index << 4) + int16(chunk.r[0])
}

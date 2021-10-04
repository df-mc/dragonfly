package chunk

import (
	"bytes"
	"github.com/df-mc/dragonfly/server/block/cube"
	"sync"
)

const (
	MaxSubChunkIndex = (cube.MaxY >> 4) - minSubChunkY

	minSubChunkY  = cube.MinY >> 4
	subChunkCount = MaxSubChunkIndex + 1
)

// Chunk is a segment in the world with a size of 16x16x256 blocks. A chunk contains multiple sub chunks
// and stores other information such as biomes.
// It is not safe to call methods on Chunk simultaneously from multiple goroutines.
type Chunk struct {
	sync.Mutex
	// air is the runtime ID of air.
	air uint32
	// sub holds all sub chunks part of the chunk. The pointers held by the array are nil if no sub chunk is
	// allocated at the indices.
	sub [subChunkCount]*SubChunk
	// biomes is an array of biome IDs. There is one biome ID for every column in the chunk.
	biomes [256]uint8
	// blockEntities holds all block entities of the chunk, prefixed by their absolute position.
	blockEntities map[cube.Pos]map[string]interface{}
}

// New initialises a new chunk and returns it, so that it may be used.
func New(airRuntimeID uint32) *Chunk {
	return &Chunk{air: airRuntimeID, blockEntities: make(map[cube.Pos]map[string]interface{})}
}

// Sub returns a list of all sub chunks present in the chunk.
func (chunk *Chunk) Sub() []*SubChunk {
	return chunk.sub[:]
}

// BiomeID returns the biome ID at a specific column in the chunk.
func (chunk *Chunk) BiomeID(x, z uint8) uint8 {
	return chunk.biomes[columnOffset(x, z)]
}

// SetBiomeID sets the biome ID at a specific column in the chunk.
func (chunk *Chunk) SetBiomeID(x, z, biomeID uint8) {
	chunk.biomes[columnOffset(x, z)] = biomeID
}

// Light returns the light level at a specific position in the chunk.
func (chunk *Chunk) Light(x uint8, y int16, z uint8) uint8 {
	sub := chunk.subChunk(y)
	if sub == nil {
		return 15
	}
	return sub.Light(x&15, uint8(y&15), z&15)
}

// SkyLight returns the sky light level at a specific position in the chunk.
func (chunk *Chunk) SkyLight(x uint8, y int16, z uint8) uint8 {
	sub := chunk.subChunk(y)
	if sub == nil {
		return 15
	}
	return sub.SkyLightAt(x&15, uint8(y&15), z&15)
}

// RuntimeID returns the runtime ID of the block at a given x, y and z in a chunk at the given layer. If no
// sub chunk exists at the given y, the block is assumed to be air.
func (chunk *Chunk) RuntimeID(x uint8, y int16, z uint8, layer uint8) uint32 {
	sub := chunk.subChunk(y)
	if sub == nil {
		// The sub chunk was not initialised, so we can conclude that the block at that location would be
		// an air block.
		return chunk.air
	}
	if uint8(len(sub.storages)) <= layer {
		return sub.air
	}
	return sub.storages[layer].RuntimeID(x, uint8(y), z)
}

// fullSkyLight is used to copy full light to newly created sub chunks.
var fullSkyLight [2048]byte

func init() {
	b := bytes.Repeat([]byte{0xff}, 2048)
	copy(fullSkyLight[:], b)
}

// SetRuntimeID sets the runtime ID of a block at a given x, y and z in a chunk at the given layer. If no
// SubChunk exists at the given y, a new SubChunk is created and the block is set.
func (chunk *Chunk) SetRuntimeID(x uint8, y int16, z uint8, layer uint8, runtimeID uint32) {
	i := subIndex(y)
	sub := chunk.sub[i]
	if sub == nil {
		// The first layer is initialised in the next call to Layer().
		sub = NewSubChunk(chunk.air)
		sub.skyLight = fullSkyLight
		chunk.sub[i] = sub
	}
	if len(sub.storages) < int(layer+1) && runtimeID == chunk.air {
		// Air was set at n layer, but there were less than n layers, so there already was air there.
		// Don't do anything with this, just return.
		return
	}
	sub.Layer(layer).SetRuntimeID(x, uint8(y), z, runtimeID)
}

// HighestLightBlocker iterates from the highest non-empty sub chunk downwards to find the Y value of the
// highest block that completely blocks any light from going through. If none is found, the value returned is
// 0.
func (chunk *Chunk) HighestLightBlocker(x, z uint8) int16 {
	for index := int16(MaxSubChunkIndex); index >= 0; index-- {
		sub := chunk.sub[index]
		if sub == nil || len(sub.storages) == 0 {
			continue
		}
		for y := 15; y >= 0; y-- {
			totalY := int16(y) | subY(index)
			if FilteringBlocks[sub.storages[0].RuntimeID(x, uint8(totalY), z)] == 15 {
				return totalY
			}
		}
	}
	return cube.MinY
}

// HighestBlock iterates from the highest non-empty sub chunk downwards to find the Y value of the highest
// non-air block at an x and z. If no blocks are present in the column, 0 is returned.
func (chunk *Chunk) HighestBlock(x, z uint8) int16 {
	for index := int16(MaxSubChunkIndex); index >= 0; index-- {
		sub := chunk.sub[index]
		if sub == nil || len(sub.storages) == 0 {
			continue
		}
		for y := 15; y >= 0; y-- {
			totalY := int16(y) | subY(index)
			rid := sub.storages[0].RuntimeID(x, uint8(totalY), z)
			if rid != chunk.air {
				return totalY
			}
		}
	}
	return cube.MinY
}

// SetBlockNBT sets block NBT data to a given position in the chunk. If the data passed is nil, the block NBT
// currently present will be cleared.
func (chunk *Chunk) SetBlockNBT(pos cube.Pos, data map[string]interface{}) {
	if data == nil {
		delete(chunk.blockEntities, pos)
		return
	}
	chunk.blockEntities[pos] = data
}

// BlockNBT returns a list of all block NBT data set in the chunk.
func (chunk *Chunk) BlockNBT() map[cube.Pos]map[string]interface{} {
	return chunk.blockEntities
}

// Compact compacts the chunk as much as possible, getting rid of any sub chunks that are empty, and compacts
// all storages in the sub chunks to occupy as little space as possible.
// Compact should be called right before the chunk is saved in order to optimise the storage space.
func (chunk *Chunk) Compact() {
	for i, sub := range chunk.sub {
		if sub == nil {
			continue
		}
		sub.compact()
		if len(sub.storages) == 0 {
			chunk.sub[i] = nil
		}
	}
}

// columnOffset returns the offset in a byte slice that the column at a specific x and z may be found.
func columnOffset(x, z uint8) uint8 {
	return (x & 15) | (z&15)<<4
}

// subChunk finds the correct SubChunk in the Chunk by a Y value.
func (chunk *Chunk) subChunk(y int16) *SubChunk {
	i := subIndex(y)
	return chunk.sub[i]
}

// subIndex returns the sub chunk Y index matching the y value passed.
func subIndex(y int16) int16 {
	return (y >> 4) - minSubChunkY
}

// subY returns the sub chunk Y value matching the index passed.
func subY(index int16) int16 {
	return (index + minSubChunkY) << 4
}

package chunk

import (
	"sync"
)

// chunk is a segment in the world with a size of 16x16x256 blocks. A chunk contains multiple sub chunks
// and stores other information such as biomes.
// It is not safe to call methods on Chunk simultaneously from multiple goroutines.
type Chunk struct {
	sync.RWMutex
	// sub holds all sub chunks part of the chunk. The pointers held by the array are nil if no sub chunk is
	// allocated at the indices.
	sub [16]*SubChunk
	// biomes is an array of biome IDs. There is one biome ID for every column in the chunk.
	biomes [256]uint8
	// blockEntities holds all block entities of the chunk, prefixed by their absolute position.
	blockEntities map[[3]int]map[string]interface{}
}

// New initialises a new chunk and returns it, so that it may be used.
func New() *Chunk {
	return &Chunk{blockEntities: make(map[[3]int]map[string]interface{})}
}

// Sub returns the sub chunk at the given y, which is a number from 0-15. If the number is higher than that,
// the y value will 'overflow' and return the sub chunk at y % 16.
// If a sub chunk at this y is not present, nil is returned.
func (chunk *Chunk) Sub(y uint8) *SubChunk {
	return chunk.sub[y&15]
}

// BiomeID returns the biome ID at a specific column in the chunk.
func (chunk *Chunk) BiomeID(x, z uint8) uint8 {
	return chunk.biomes[columnOffset(x, z)]
}

// SetBiomeID sets the biome ID at a specific column in the chunk.
func (chunk *Chunk) SetBiomeID(x, z, biomeID uint8) {
	chunk.biomes[columnOffset(x, z)] = biomeID
}

// BlockRuntimeID returns the runtime ID of the block at a given x, y and z in a chunk at the given layer. If no
// sub chunk exists at the given y, the block is assumed to be air.
func (chunk *Chunk) RuntimeID(x, y, z uint8, layer uint8) uint32 {
	subChunkY := y >> 4
	for i := byte(0); i <= subChunkY; i++ {
		// The sub chunk was not initialised, so we can conclude that the block at that location would be
		// an air block. (always runtime ID 0)
		if chunk.sub[i] == nil {
			return 0
		}
	}
	return chunk.sub[subChunkY].Layer(layer).RuntimeID(x, y, z)
}

// SetRuntimeID sets the runtime ID of a block at a given x, y and z in a chunk at the given layer. If no
// SubChunk exists at the given y, a new SubChunk is created and the block is set.
func (chunk *Chunk) SetRuntimeID(x, y, z uint8, layer uint8, runtimeID uint32) {
	i := y >> 4
	if chunk.sub[i] == nil {
		chunk.sub[i] = &SubChunk{}
		// Initialise the first layer of the SubChunk.
		chunk.sub[i].Layer(layer)
	}
	chunk.sub[i].Layer(layer).SetRuntimeID(x, y, z, runtimeID)
}

// SetBlockNBT sets block NBT data to a given position in the chunk. If the data passed is nil, the block NBT
// currently present will be cleared.
func (chunk *Chunk) SetBlockNBT(pos [3]int, data map[string]interface{}) {
	if data == nil {
		delete(chunk.blockEntities, pos)
		return
	}
	chunk.blockEntities[pos] = data
}

// BlockNBT returns a list of all block NBT data set in the chunk.
func (chunk *Chunk) BlockNBT() map[[3]int]map[string]interface{} {
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

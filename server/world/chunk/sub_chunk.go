package chunk

// SubChunk is a cube of blocks located in a chunk. It has a size of 16x16x16 blocks and forms part of a stack
// that forms a Chunk.
type SubChunk struct {
	air        uint32
	storages   []*BlockStorage
	blockLight [2048]uint8
	skyLight   [2048]uint8
}

// NewSubChunk creates a new sub chunk. All sub chunks should be created through this function
func NewSubChunk(airRuntimeID uint32) *SubChunk {
	return &SubChunk{air: airRuntimeID}
}

// Layer returns a certain block storage/layer from a sub chunk. If no storage at the layer exists, the layer
// is created, as well as all layers between the current highest layer and the new highest layer.
func (sub *SubChunk) Layer(layer uint8) *BlockStorage {
	for i := uint8(len(sub.storages)); i <= layer; i++ {
		// Keep appending to storages until the requested layer is achieved. Makes working with new layers
		// much easier.
		sub.addLayer()
	}
	return sub.storages[layer]
}

// addLayer adds a new storage at the next layer. This is forced to not inline to guarantee that Layer is
// inlined.
//go:noinline
func (sub *SubChunk) addLayer() {
	sub.storages = append(sub.storages, newBlockStorage(make([]uint32, 128), newPalette(1, []uint32{sub.air})))
}

// Layers returns all layers in the sub chunk. This method may also return an empty slice.
func (sub *SubChunk) Layers() []*BlockStorage {
	return sub.storages
}

// RuntimeID returns the runtime ID of the block located at the given X, Y and Z. X, Y and Z must be in a
// range of 0-15.
func (sub *SubChunk) RuntimeID(x, y, z byte, layer uint8) uint32 {
	if uint8(len(sub.storages)) <= layer {
		return sub.air
	}
	return sub.storages[layer].RuntimeID(x, y, z)
}

// SetRuntimeID sets the given runtime ID at the given X, Y and Z. X, Y and Z must be in a range of 0-15.
func (sub *SubChunk) SetRuntimeID(x, y, z byte, layer uint8, runtimeID uint32) {
	sub.Layer(layer).SetRuntimeID(x, y, z, runtimeID)
}

// Light returns the light level at a specific position in the sub chunk. It is max(block light, sky light).
func (sub *SubChunk) Light(x, y, z byte) uint8 {
	skyLight := sub.SkyLightAt(x, y, z)
	if skyLight == 15 {
		// The sky light was already on the maximum value, so return it with checking block light.
		return 15
	}
	blockLight := sub.blockLightAt(x, y, z)
	if skyLight > blockLight {
		return skyLight
	}
	return blockLight
}

// ClearLight clears all light of the sub chunk.
func (sub *SubChunk) ClearLight() {
	sub.skyLight = [2048]uint8{}
	sub.blockLight = [2048]uint8{}
}

// setBlockLight sets the block light value at a specific position in the sub chunk.
func (sub *SubChunk) setBlockLight(x, y, z byte, level uint8) {
	index := (uint16(x) << 8) | (uint16(z) << 4) | uint16(y)

	i := index >> 1
	bit := (index & 1) << 2
	sub.blockLight[i] = (sub.blockLight[i] & (0xf0 >> bit)) | (level << bit)
}

// blockLightAt returns the block light value at a specific value at a specific position in the sub chunk.
func (sub *SubChunk) blockLightAt(x, y, z byte) uint8 {
	index := (uint16(x) << 8) | (uint16(z) << 4) | uint16(y)

	return (sub.blockLight[index>>1] >> ((index & 1) << 2)) & 0xf
}

// setSkyLight sets the sky light value at a specific position in the sub chunk.
func (sub *SubChunk) setSkyLight(x, y, z byte, level uint8) {
	index := (uint16(x) << 8) | (uint16(z) << 4) | uint16(y)

	i := index >> 1
	bit := (index & 1) << 2
	sub.skyLight[i] = (sub.skyLight[i] & (0xf0 >> bit)) | (level << bit)
}

// SkyLightAt returns the sky light value at a specific value at a specific position in the sub chunk.
func (sub *SubChunk) SkyLightAt(x, y, z byte) uint8 {
	index := (uint16(x) << 8) | (uint16(z) << 4) | uint16(y)

	return (sub.skyLight[index>>1] >> ((index & 1) << 2)) & 0xf
}

// Compact cleans the garbage from all block storages that sub chunk contains, so that they may be
// cleanly written to a database.
func (sub *SubChunk) compact() {
	newStorages := make([]*BlockStorage, 0, len(sub.storages))
	for _, storage := range sub.storages {
		storage.compact()
		if len(storage.palette.blockRuntimeIDs) == 1 && storage.palette.blockRuntimeIDs[0] == sub.air {
			// If the palette has only air in it, it means the storage is empty, so we can ignore it.
			continue
		}
		newStorages = append(newStorages, storage)
	}
	sub.storages = newStorages
}

package chunk

// SubChunk is a cube of blocks located in a chunk. It has a size of 16x16x16 blocks and forms part of a stack
// that forms a Chunk.
type SubChunk struct {
	storages   []*BlockStorage
	blockLight [2048]uint8
	skyLight   [2048]uint8
}

// Layer returns a certain block storage/layer from a sub chunk. If no storage at the layer exists, the layer
// is created, as well as all layers between the current highest layer and the new highest layer.
func (sub *SubChunk) Layer(layer uint8) *BlockStorage {
	for uint8(len(sub.storages)) <= layer {
		// Keep appending to storages until the requested layer is achieved. Makes working with new layers
		// much easier.
		sub.storages = append(sub.storages, newBlockStorage(make([]uint32, 128), newPalette(1, []uint32{0})))
	}
	return sub.storages[layer]
}

// Layers returns all layers in the sub chunk. This method may also return an empty slice.
func (sub *SubChunk) Layers() []*BlockStorage {
	return sub.storages
}

// RuntimeID returns the runtime ID of the block located at the given X, Y and Z. X, Y and Z must be in a
// range of 0-15.
func (sub *SubChunk) RuntimeID(x, y, z byte, layer uint8) uint32 {
	if uint8(len(sub.storages)) <= layer {
		return 0
	}
	return sub.Layer(layer).RuntimeID(x, y, z)
}

// SetRuntimeID sets the given runtime ID at the given X, Y and Z. X, Y and Z must be in a range of 0-15.
func (sub *SubChunk) SetRuntimeID(x, y, z byte, layer uint8, runtimeID uint32) {
	sub.Layer(layer).SetRuntimeID(x, y, z, runtimeID)
}

// Light returns the light level at a specific position in the sub chunk. It is max(block light, sky light).
func (sub *SubChunk) Light(x, y, z byte) uint8 {
	skyLight := sub.skyLightAt(x, y, z)
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

// setBlockLight sets the block light value at a specific position in the sub chunk.
func (sub *SubChunk) setBlockLight(x, y, z byte, level uint8) {
	index := (uint16(x) << 8) | (uint16(z) << 4) | uint16(y)
	i := index >> 1
	bit := index & 1
	sub.blockLight[i] = ((0xF << uint(bit<<2)) & sub.blockLight[i]) | ((level & 0xf) << uint((1^bit)<<2))
}

// blockLightAt returns the block light value at a specific value at a specific position in the sub chunk.
func (sub *SubChunk) blockLightAt(x, y, z byte) uint8 {
	index := (uint16(x) << 8) | (uint16(z) << 4) | uint16(y)
	i := index >> 1
	if index&1 == 0 {
		return sub.blockLight[i] >> 4
	}
	return sub.blockLight[i] & 0xf
}

// setSkyLight sets the sky light value at a specific position in the sub chunk.
func (sub *SubChunk) setSkyLight(x, y, z byte, level uint8) {
	index := (uint16(x) << 8) | (uint16(z) << 4) | uint16(y)

	i := index >> 1
	bit := index & 1
	sub.skyLight[i] = ((0xF << uint(bit<<2)) & sub.skyLight[i]) | ((level & 0xf) << uint((1^bit)<<2))
}

// skyLightAt returns the sky light value at a specific value at a specific position in the sub chunk.
func (sub *SubChunk) skyLightAt(x, y, z byte) uint8 {
	index := (uint16(x) << 8) | (uint16(z) << 4) | uint16(y)
	i := index >> 1
	if index&1 == 0 {
		return sub.skyLight[i] >> 4
	}
	return sub.skyLight[i] & 0xf
}

// Compact cleans the garbage from all block storages that sub chunk contains, so that they may be
// cleanly written to a database.
func (sub *SubChunk) compact() {
	newStorages := make([]*BlockStorage, 0, len(sub.storages))
	for _, storage := range sub.storages {
		storage.compact()
		if len(storage.palette.blockRuntimeIDs) == 1 && storage.palette.blockRuntimeIDs[0] == 0 {
			// If the palette has only air in it, it means the storage is empty, so we can ignore it.
			continue
		}
		newStorages = append(newStorages, storage)
	}
	sub.storages = newStorages
}

package chunk

import "git.jetbrains.space/dragonfly/dragonfly/dragonfly/block"

// SubChunk is a cube of blocks located in a chunk. It has a size of 16x16x16 blocks and forms part of a stack
// that forms a Chunk.
type SubChunk struct {
	storages []*BlockStorage
}

// Layer returns a certain block storage/layer from a sub chunk. If no storage at the layer exists, the layer
// is created, as well as all layers between the current highest layer and the new highest layer.
func (subChunk *SubChunk) Layer(layer uint8) *BlockStorage {
	for uint8(len(subChunk.storages)) <= layer {
		// Keep appending to storages until the requested layer is achieved. Makes working with new layers
		// much easier.
		id, _ := block.RuntimeID(block.Air{})
		subChunk.storages = append(subChunk.storages, newBlockStorage(make([]uint32, 128), newPalette(1, []uint32{id})))
	}
	return subChunk.storages[layer]
}

// RuntimeID returns the runtime ID of the block located at the given X, Y and Z. X, Y and Z must be in a
// range of 0-15.
func (subChunk *SubChunk) RuntimeID(x, y, z byte, layer uint8) uint32 {
	return subChunk.Layer(layer).RuntimeID(x, y, z)
}

// SetRuntimeID sets the given runtime ID at the given X, Y and Z. X, Y and Z must be in a range of 0-15.
func (subChunk *SubChunk) SetRuntimeID(x, y, z byte, layer uint8, runtimeID uint32) {
	subChunk.Layer(layer).SetRuntimeID(x, y, z, runtimeID)
}

// Compact cleans the garbage from all block storages that sub chunk contains, so that they may be
// cleanly written to a database.
func (subChunk *SubChunk) compact() {
	id, _ := block.RuntimeID(block.Air{})
	newStorages := make([]*BlockStorage, 0, len(subChunk.storages))
	for _, storage := range subChunk.storages {
		storage.compact()
		if len(storage.palette.blockRuntimeIDs) == 1 && storage.palette.blockRuntimeIDs[0] == id {
			// If the palette has only air in it, it means the storage is empty, so we can ignore it.
			continue
		}
		newStorages = append(newStorages, storage)
	}
	subChunk.storages = newStorages
}

package chunk

const (
	// uint32ByteSize is the amount of bytes in a uint32.
	uint32ByteSize = 4
	// uint32BitSize is the amount of bits in a uint32.
	uint32BitSize = uint32ByteSize * 8
)

// BlockStorage is a storage of 4096 blocks encoded in a variable amount of uint32s, storages may have blocks
// with a bit size per block of 1, 2, 3, 4, 5, 6, 8 or 16 bits.
// 3 of these formats have additional padding in every uint32 and an additional uint32 at the end, to cater
// for the blocks that don't fit. This padding is present when the storage has a block size of 3, 5 or 6
// bytes.
// Methods on BlockStorage must not be called simultaneously from multiple goroutines.
type BlockStorage struct {
	// bitsPerBlock is the amount of bits required to store one block. The number increases as the block
	// storage holds more unique block states.
	bitsPerBlock uint16
	blockMask    uint32
	// filledBitsPerWord returns the amount of blocks that are actually filled per uint32.
	filledBitsPerWord uint16
	// Palette holds all block runtime IDs that the blocks in the blocks slice point to. These runtime IDs
	// point to block states.
	palette *Palette

	// blocks contains all blocks in the block storage. This slice has a variable size, but may not be changed
	// unless the whole block storage is resized, including the palette.
	blocks []uint32
}

// NewBlockStorage creates a new block storage using the uint32 slice as the blocks and the palette passed.
// The bits per block are calculated using the length of the uint32 slice.
func newBlockStorage(blocks []uint32, palette *Palette) *BlockStorage {
	bitsPerBlock := uint16(len(blocks) / uint32BitSize / uint32ByteSize)
	return &BlockStorage{blocks: blocks, bitsPerBlock: bitsPerBlock, blockMask: (1 << bitsPerBlock) - 1, filledBitsPerWord: uint32BitSize / bitsPerBlock * bitsPerBlock, palette: palette}
}

// Palette returns the Palette of the block storage.
func (storage *BlockStorage) Palette() *Palette {
	return storage.palette
}

// RuntimeID returns the runtime ID of the block located at the given x, y and z.
func (storage *BlockStorage) RuntimeID(x, y, z byte) uint32 {
	return storage.palette.RuntimeID(storage.paletteOffset(x&15, y&15, z&15))
}

// SetRuntimeID sets the given runtime ID at the given x, y and z. The palette and block storage are expanded
// automatically to make space for the runtime ID, should that be needed.
func (storage *BlockStorage) SetRuntimeID(x, y, z byte, runtimeID uint32) {
	index := storage.palette.Index(runtimeID)
	if index == -1 {
		// The runtime ID was not yet available in the palette. We add it, then check if the block storage
		// needs to be resized for the palette pointers to fit.
		index = int(storage.palette.Add(runtimeID))

		if storage.palette.needsResize() {
			storage.palette.increaseSize()
			storage.resize(storage.palette.size)
		}
	}
	storage.setPaletteOffset(x&15, y&15, z&15, uint16(index))
}

// paletteOffset looks up the palette offset at a given x, y and z value in the block storage. This palette
// offset is not the runtime ID at this offset, but merely an offset in the palette, pointing to a runtime ID.
func (storage *BlockStorage) paletteOffset(x, y, z byte) uint16 {
	offset := ((uint16(x) << 8) | (uint16(z) << 4) | uint16(y)) * storage.bitsPerBlock
	uint32Offset, bitOffset := offset/storage.filledBitsPerWord, offset%storage.filledBitsPerWord
	return uint16((storage.blocks[uint32Offset] >> bitOffset) & storage.blockMask)
}

// setPaletteOffset sets the palette offset at a given x, y and z to paletteOffset. This offset should point
// to a runtime ID in the block storage's palette.
func (storage *BlockStorage) setPaletteOffset(x, y, z byte, paletteOffset uint16) {
	offset := ((uint16(x) << 8) | (uint16(z) << 4) | uint16(y)) * storage.bitsPerBlock
	uint32Offset, bitOffset := offset/storage.filledBitsPerWord, offset%storage.filledBitsPerWord

	storage.blocks[uint32Offset] = storage.blocks[uint32Offset]&^(storage.blockMask<<bitOffset) | uint32(paletteOffset<<bitOffset)
}

// resize changes the size of a block storage to palette size newPaletteSize. A new block storage is
// constructed, and all blocks available in the current storage are set in their appropriate locations in the
// new storage.
func (storage *BlockStorage) resize(newPaletteSize paletteSize) {
	if newPaletteSize == paletteSize(storage.bitsPerBlock) {
		return // Don't resize if the size is already equal.
	}

	const subChunkBlockCount = 16 * 16 * 16
	requiredUint32s := subChunkBlockCount / int(uint32BitSize/newPaletteSize)
	if newPaletteSize == 3 || newPaletteSize == 5 || newPaletteSize == 6 {
		// Add one uint32 if the palette size is one of the padded sizes.
		requiredUint32s++
	}
	n := make([]uint32, requiredUint32s)

	// Construct a new block storage, set all blocks in there manually. We can't easily do this in a better
	// way, because all blocks will be at a different offset with a different length.
	newStorage := newBlockStorage(n, storage.palette)
	for x := byte(0); x < 16; x++ {
		for y := byte(0); y < 16; y++ {
			for z := byte(0); z < 16; z++ {
				newStorage.setPaletteOffset(x, y, z, storage.paletteOffset(x, y, z))
			}
		}
	}
	// Set the new storage.
	*storage = *newStorage
}

// compact clears unused indexes in the palette by scanning for usages in the block storage. This is a
// relatively heavy task which should only happen right before the sub chunk holding this block storage is
// saved to disk.
func (storage *BlockStorage) compact() {
	usedIndices := make([]bool, storage.palette.Len())
	for x := byte(0); x < 16; x++ {
		for y := byte(0); y < 16; y++ {
			for z := byte(0); z < 16; z++ {
				usedIndices[storage.paletteOffset(x, y, z)] = true
			}
		}
	}
	newRuntimeIDs := make([]uint32, 0, len(usedIndices))
	conversion := make([]uint16, len(usedIndices))

	for index, set := range usedIndices {
		if set {
			conversion[index] = uint16(len(newRuntimeIDs))
			newRuntimeIDs = append(newRuntimeIDs, storage.palette.blockRuntimeIDs[index])
		}
	}
	for x := byte(0); x < 16; x++ {
		for y := byte(0); y < 16; y++ {
			for z := byte(0); z < 16; z++ {
				// Replace all usages of the old palette indexes with the new indexes using the map we
				// produced earlier.
				storage.setPaletteOffset(x, y, z, conversion[storage.paletteOffset(x, y, z)])
			}
		}
	}
	storage.palette.blockRuntimeIDs = newRuntimeIDs
}

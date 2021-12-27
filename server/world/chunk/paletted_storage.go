package chunk

import (
	"reflect"
	"unsafe"
)

const (
	// uint32ByteSize is the amount of bytes in a uint32.
	uint32ByteSize = 4
	// uint32BitSize is the amount of bits in a uint32.
	uint32BitSize = uint32ByteSize * 8
)

// PalettedStorage is a storage of 4096 blocks encoded in a variable amount of uint32s, storages may have values
// with a bit size per block of 0, 1, 2, 3, 4, 5, 6, 8 or 16 bits.
// 3 of these formats have additional padding in every uint32 and an additional uint32 at the end, to cater
// for the blocks that don't fit. This padding is present when the storage has a block size of 3, 5 or 6
// bytes.
// Methods on PalettedStorage must not be called simultaneously from multiple goroutines.
type PalettedStorage struct {
	// bitsPerIndex is the amount of bits required to store one block. The number increases as the block
	// storage holds more unique block states.
	bitsPerIndex uint16
	// filledBitsPerIndex returns the amount of blocks that are actually filled per uint32.
	filledBitsPerIndex uint16
	// indexMask is the equivalent of 1 << bitsPerIndex - 1.
	indexMask uint32

	// indicesStart holds an unsafe.Pointer to the first byte in the indices slice below.
	indicesStart unsafe.Pointer

	// Palette holds all block runtime IDs that the indices in the indices slice point to. These runtime IDs
	// point to block states.
	palette *Palette

	// indices contains all indices in the PalettedStorage. This slice has a variable size, but may not be changed
	// unless the whole PalettedStorage is resized, including the Palette.
	indices []uint32
}

// newPalettedStorage creates a new block storage using the uint32 slice as the indices and the palette passed.
// The bits per block are calculated using the length of the uint32 slice.
func newPalettedStorage(indices []uint32, palette *Palette) *PalettedStorage {
	var (
		bitsPerIndex       = uint16(len(indices) / uint32BitSize / uint32ByteSize)
		indexMask          = (uint32(1) << bitsPerIndex) - 1
		indicesStart       = (unsafe.Pointer)((*reflect.SliceHeader)(unsafe.Pointer(&indices)).Data)
		filledBitsPerIndex uint16
	)
	if bitsPerIndex != 0 {
		filledBitsPerIndex = uint32BitSize / bitsPerIndex * bitsPerIndex
	}
	return &PalettedStorage{filledBitsPerIndex: filledBitsPerIndex, indexMask: indexMask, indicesStart: indicesStart, bitsPerIndex: bitsPerIndex, indices: indices, palette: palette}
}

// emptyStorage creates a PalettedStorage filled completely with a value v.
func emptyStorage(v uint32) *PalettedStorage {
	return newPalettedStorage([]uint32{}, newPalette(0, []uint32{v}))
}

// Palette returns the Palette of the PalettedStorage.
func (storage *PalettedStorage) Palette() *Palette {
	return storage.palette
}

// At returns the value of the PalettedStorage at a given x, y and z.
func (storage *PalettedStorage) At(x, y, z byte) uint32 {
	return storage.palette.Value(storage.paletteIndex(x&15, y&15, z&15))
}

// Set sets a value at a specific x, y and z. The Palette and PalettedStorage are expanded
// automatically to make space for the value, should that be needed.
func (storage *PalettedStorage) Set(x, y, z byte, v uint32) {
	index := storage.palette.Index(v)
	if index == -1 {
		// The runtime ID was not yet available in the palette. We add it, then check if the block storage
		// needs to be resized for the palette pointers to fit.
		index = storage.addNew(v)
	}
	storage.setPaletteIndex(x&15, y&15, z&15, uint16(index))
}

// addNew adds a new value to the PalettedStorage's Palette and returns its index. If needed, the storage is resized.
func (storage *PalettedStorage) addNew(v uint32) int16 {
	index, resize := storage.palette.Add(v)
	if resize {
		storage.resize(storage.palette.size)
	}
	return index
}

// paletteIndex looks up the Palette index at a given x, y and z value in the PalettedStorage. This palette
// index is not the value at this offset, but merely an index in the Palette pointing to a value.
func (storage *PalettedStorage) paletteIndex(x, y, z byte) uint16 {
	if storage.bitsPerIndex == 0 {
		// Unfortunately our default logic cannot deal with 0 bits per index, meaning we'll have to special case
		// this. This comes with a little performance hit, but it seems to be the only way to go. An alternative would
		// be not to have 0 bits per block storages in memory, but that would cause a strongly increased memory usage
		// by biomes.
		return 0
	}
	offset := ((uint16(x) << 8) | (uint16(z) << 4) | uint16(y)) * storage.bitsPerIndex
	uint32Offset, bitOffset := offset/storage.filledBitsPerIndex, offset%storage.filledBitsPerIndex

	w := *(*uint32)(unsafe.Pointer(uintptr(storage.indicesStart) + uintptr(uint32Offset<<2)))
	return uint16((w >> bitOffset) & storage.indexMask)
}

// setPaletteIndex sets the palette index at a given x, y and z to paletteIndex. This index should point
// to a value in the PalettedStorage's Palette.
func (storage *PalettedStorage) setPaletteIndex(x, y, z byte, i uint16) {
	if storage.bitsPerIndex == 0 {
		return
	}
	offset := ((uint16(x) << 8) | (uint16(z) << 4) | uint16(y)) * storage.bitsPerIndex
	uint32Offset, bitOffset := offset/storage.filledBitsPerIndex, offset%storage.filledBitsPerIndex

	ptr := (*uint32)(unsafe.Pointer(uintptr(storage.indicesStart) + uintptr(uint32Offset<<2)))
	*ptr = (*ptr &^ (storage.indexMask << bitOffset)) | (uint32(i) << bitOffset)
}

// resize changes the size of a PalettedStorage to newPaletteSize. A new PalettedStorage is constructed,
// and all values available in the current storage are set in their appropriate locations in the
// new storage.
func (storage *PalettedStorage) resize(newPaletteSize paletteSize) {
	if newPaletteSize == paletteSize(storage.bitsPerIndex) {
		return // Don't resize if the size is already equal.
	}
	// Construct a new storage and set all values in there manually. We can't easily do this in a better
	// way, because all values will be at a different index with a different length.
	newStorage := newPalettedStorage(make([]uint32, newPaletteSize.uint32s()), storage.palette)
	for x := byte(0); x < 16; x++ {
		for y := byte(0); y < 16; y++ {
			for z := byte(0); z < 16; z++ {
				newStorage.setPaletteIndex(x, y, z, storage.paletteIndex(x, y, z))
			}
		}
	}
	// Set the new storage.
	*storage = *newStorage
}

// compact clears unused indexes in the palette by scanning for usages in the PalettedStorage. This is a
// relatively heavy task which should only happen right before the sub chunk holding this PalettedStorage is
// saved to disk. compact also shrinks the palette size if possible.
func (storage *PalettedStorage) compact() {
	usedIndices := make([]bool, storage.palette.Len())
	for x := byte(0); x < 16; x++ {
		for y := byte(0); y < 16; y++ {
			for z := byte(0); z < 16; z++ {
				usedIndices[storage.paletteIndex(x, y, z)] = true
			}
		}
	}
	newRuntimeIDs := make([]uint32, 0, len(usedIndices))
	conversion := make([]uint16, len(usedIndices))

	for index, set := range usedIndices {
		if set {
			conversion[index] = uint16(len(newRuntimeIDs))
			newRuntimeIDs = append(newRuntimeIDs, storage.palette.values[index])
		}
	}
	// Construct a new storage and set all values in there manually. We can't easily do this in a better
	// way, because all values will be at a different index with a different length.
	size := paletteSizeFor(len(newRuntimeIDs))
	newStorage := newPalettedStorage(make([]uint32, size.uint32s()), newPalette(size, newRuntimeIDs))

	for x := byte(0); x < 16; x++ {
		for y := byte(0); y < 16; y++ {
			for z := byte(0); z < 16; z++ {
				// Replace all usages of the old palette indexes with the new indexes using the map we
				// produced earlier.
				newStorage.setPaletteIndex(x, y, z, conversion[storage.paletteIndex(x, y, z)])
			}
		}
	}
	*storage = *newStorage
}

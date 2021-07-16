package chunk

import "math"

// paletteSize is the size of a palette. It indicates the amount of bits occupied per block saved.
type paletteSize byte

// Palette is a palette of runtime IDs that every block storage has. Block storages hold 'pointers' to indexes
// in this palette.
type Palette struct {
	last      uint32
	lastIndex int16
	size      paletteSize

	// blockRuntimeIDs is a map of runtime IDs. The block storages point to the index to this runtime ID.
	blockRuntimeIDs []uint32
}

// newPalette returns a new palette with size and a slice of added runtime IDs.
func newPalette(size paletteSize, runtimeIDs []uint32) *Palette {
	return &Palette{size: size, blockRuntimeIDs: runtimeIDs, last: math.MaxUint32}
}

// Len returns the amount of unique block runtime IDs in the palette.
func (palette *Palette) Len() int {
	return len(palette.blockRuntimeIDs)
}

// Add adds a runtime ID to the palette. It does not first if the runtime ID was already set in the palette.
// The index at which the runtime ID was added is returned.
// Another bool is returned indicating if the palette was resized as a result of the adding of the runtime ID.
func (palette *Palette) Add(runtimeID uint32) (index int16, resize bool) {
	i := int16(len(palette.blockRuntimeIDs))
	palette.blockRuntimeIDs = append(palette.blockRuntimeIDs, runtimeID)

	if palette.needsResize() {
		palette.increaseSize()
		return i, true
	}
	return i, false
}

// Replace calls the function passed for each runtime ID present in the palette. The value returned by the
// function replaces the runtime ID present at the index of the runtime ID passed.
func (palette *Palette) Replace(f func(runtimeID uint32) uint32) {
	// Reset last runtime ID as it now has a different offset.
	palette.last = math.MaxUint32
	for index, id := range palette.blockRuntimeIDs {
		palette.blockRuntimeIDs[index] = f(id)
	}
}

// Index loops through the runtime IDs of the palette and looks for the index of the given runtime ID. If the
// runtime ID can not be found, -1 is returned.
func (palette *Palette) Index(runtimeID uint32) int16 {
	if runtimeID == palette.last {
		// Fast path out.
		return palette.lastIndex
	}
	// Slow path in a separate function allows for inlining the fast path.
	return palette.indexSlow(runtimeID)
}

// indexSlow searches the index of a runtime ID in the palette's block runtime IDs by iterating through the
// palette's block runtime IDs.
func (palette *Palette) indexSlow(runtimeID uint32) int16 {
	l := len(palette.blockRuntimeIDs)
	for i := 0; i < l; i++ {
		if palette.blockRuntimeIDs[i] == runtimeID {
			palette.last = runtimeID
			v := int16(i)
			palette.lastIndex = v
			return v
		}
	}
	return -1
}

// RuntimeID returns the runtime ID at the palette index given.
func (palette *Palette) RuntimeID(paletteIndex uint16) uint32 {
	return palette.blockRuntimeIDs[paletteIndex]
}

// needsResize checks if the palette and with it the holding block storage needs to be resized to a bigger
// size.
func (palette *Palette) needsResize() bool {
	return len(palette.blockRuntimeIDs) > (1 << palette.size)
}

var sizes = [...]paletteSize{1, 2, 3, 4, 5, 6, 8, 16}
var offsets = [...]int{1: 0, 2: 1, 3: 2, 4: 3, 5: 4, 6: 5, 8: 6, 16: 7}

// increaseSize increases the size of the palette to the next palette size.
func (palette *Palette) increaseSize() {
	palette.size = sizes[offsets[palette.size]+1]
}

// padded returns true if the palette size is 3, 5 or 6.
func (p paletteSize) padded() bool {
	return p == 3 || p == 5 || p == 6
}

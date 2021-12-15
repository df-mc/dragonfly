package chunk

import (
	"math"
)

// paletteSize is the size of a palette. It indicates the amount of bits occupied per value stored.
type paletteSize byte

// Palette is a palette of values that every PalettedStorage has. Storages hold 'pointers' to indices
// in this palette.
type Palette struct {
	last      uint32
	lastIndex int16
	size      paletteSize

	// values is a map of values. A PalettedStorage points to the index to this value.
	values []uint32
}

// newPalette returns a new Palette with size and a slice of added values.
func newPalette(size paletteSize, values []uint32) *Palette {
	return &Palette{size: size, values: values, last: math.MaxUint32}
}

// Len returns the amount of unique values in the Palette.
func (palette *Palette) Len() int {
	return len(palette.values)
}

// Add adds a values to the Palette. It does not first check if the value was already set in the Palette.
// The index at which the value was added is returned. Another bool is returned indicating if the Palette
// was resized as a result of adding the value.
func (palette *Palette) Add(v uint32) (index int16, resize bool) {
	i := int16(len(palette.values))
	palette.values = append(palette.values, v)

	if palette.needsResize() {
		palette.increaseSize()
		return i, true
	}
	return i, false
}

// Replace calls the function passed for each value present in the Palette. The value returned by the
// function replaces the value present at the index of the value passed.
func (palette *Palette) Replace(f func(v uint32) uint32) {
	// Reset last runtime ID as it now has a different offset.
	palette.last = math.MaxUint32
	for index, v := range palette.values {
		palette.values[index] = f(v)
	}
}

// Index loops through the values of the Palette and looks for the index of the given value. If the value could
// not be found, -1 is returned.
func (palette *Palette) Index(runtimeID uint32) int16 {
	if runtimeID == palette.last {
		// Fast path out.
		return palette.lastIndex
	}
	// Slow path in a separate function allows for inlining the fast path.
	return palette.indexSlow(runtimeID)
}

// indexSlow searches the index of a value in the Palette's values by iterating through the Palette's values.
func (palette *Palette) indexSlow(runtimeID uint32) int16 {
	l := len(palette.values)
	for i := 0; i < l; i++ {
		if palette.values[i] == runtimeID {
			palette.last = runtimeID
			v := int16(i)
			palette.lastIndex = v
			return v
		}
	}
	return -1
}

// Value returns the value in the Palette at a specific index.
func (palette *Palette) Value(i uint16) uint32 {
	return palette.values[i]
}

// needsResize checks if the Palette, and with it the holding PalettedStorage, needs to be resized to a bigger
// size.
func (palette *Palette) needsResize() bool {
	return len(palette.values) > (1 << palette.size)
}

var sizes = [...]paletteSize{0, 1, 2, 3, 4, 5, 6, 8, 16}
var offsets = [...]int{0: 0, 1: 1, 2: 2, 3: 3, 4: 4, 5: 5, 6: 6, 8: 7, 16: 8}

// increaseSize increases the size of the Palette to the next palette size.
func (palette *Palette) increaseSize() {
	palette.size = sizes[offsets[palette.size]+1]
}

// padded returns true if the Palette size is 3, 5 or 6.
func (p paletteSize) padded() bool {
	return p == 3 || p == 5 || p == 6
}

// paletteSizeFor finds a suitable paletteSize for the amount of values passed n.
func paletteSizeFor(n int) paletteSize {
	for _, size := range sizes {
		if n <= (1 << size) {
			return size
		}
	}
	// Should never happen.
	return 0
}

// uint32s returns the amount of uint32s needed to represent a storage with this palette size.
func (p paletteSize) uint32s() (n int) {
	uint32Count := 0
	if p != 0 {
		// indicesPerUint32 is the amount of indices that may be stored in a single uint32.
		indicesPerUint32 := 32 / int(p)
		// uint32Count is the amount of uint32s required to store all indices: 4096 indices need to be stored in
		// total.
		uint32Count = 4096 / indicesPerUint32
	}
	if p.padded() {
		// We've got one of the padded sizes, so the storage has another uint32 to be able to store
		// every index.
		uint32Count++
	}
	return uint32Count
}

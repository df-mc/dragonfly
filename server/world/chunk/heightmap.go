package chunk

import (
	"math"
)

// HeightMap represents the height map of a chunk. It holds the y value of all the highest blocks in the chunk
// that diffuse or obstruct light.
type HeightMap []int16

// At returns the height map value at a specific column in the chunk.
func (h HeightMap) At(x, z uint8) int16 {
	return h[(uint16(x)<<4)|uint16(z)]
}

// Set changes the height map value at a specific column in the chunk.
func (h HeightMap) Set(x, z uint8, val int16) {
	h[(uint16(x)<<4)|uint16(z)] = val
}

// HighestNeighbour returns the height map value of the highest directly neighbouring column of the x and z values
// passed.
func (h HeightMap) HighestNeighbour(x, z uint8) int16 {
	highest := int16(math.MinInt16)
	if x != 15 {
		if val := h.At(x+1, z); val > highest {
			highest = val
		}
	}
	if x != 0 {
		if val := h.At(x-1, z); val > highest {
			highest = val
		}
	}
	if z != 15 {
		if val := h.At(x, z+1); val > highest {
			highest = val
		}
	}
	if z != 0 {
		if val := h.At(x, z-1); val > highest {
			highest = val
		}
	}
	return highest
}

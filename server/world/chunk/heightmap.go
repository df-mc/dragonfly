package chunk

import (
	"github.com/df-mc/dragonfly/server/block/cube"
)

// heightmap represents the heightmap of a chunk. It holds the y value of all the highest blocks in the chunk
// that diffuse or obstruct light.
type heightmap []int16

// calculateHeightmap calculates the heightmap of the chunk passed and returns it.
func calculateHeightmap(c *Chunk) heightmap {
	h := make(heightmap, 256)

	highestY := int16(cube.MinY)
	for y := int16(0); y <= maxSubChunkIndex; y++ {
		if c.sub[y] != nil {
			highestY = subY(y) + 15
		}
	}
	if highestY == cube.MinY {
		// No non-nil sub chunks present at all.
		return h
	}

	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			for y := highestY; y >= cube.MinY; y-- {
				if filterLevel(c.sub[subIndex(y)], x, uint8(y)&0xf, z) > 0 {
					h.set(x, z, y)
					break
				}
			}
		}
	}
	return h
}

// at returns the heightmap value at a specific column in the chunk.
func (h heightmap) at(x, z uint8) int16 {
	return h[(uint16(x)<<4)|uint16(z)]
}

// set sets the heightmap value at a specific column in the chunk.
func (h heightmap) set(x, z uint8, val int16) {
	h[(uint16(x)<<4)|uint16(z)] = val
}

// highestNeighbour returns the heightmap value of the highest directly neighbouring column of the x and z values
// passed.
func (h heightmap) highestNeighbour(x, z uint8) int16 {
	highest := h.at(x, z)
	if x != 15 {
		if val := h.at(x+1, z); val > highest {
			highest = val
		}
	}
	if x != 0 {
		if val := h.at(x-1, z); val > highest {
			highest = val
		}
	}
	if z != 15 {
		if val := h.at(x, z+1); val > highest {
			highest = val
		}
	}
	if z != 0 {
		if val := h.at(x, z-1); val > highest {
			highest = val
		}
	}
	return highest
}

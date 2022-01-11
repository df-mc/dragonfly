package chunk

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"math"
)

// heightmap represents the heightmap of a chunk. It holds the y value of all the highest blocks in the chunk
// that diffuse or obstruct light.
type heightmap []int16

// calculateHeightmap calculates the heightmap of the chunk passed and returns it.
func calculateHeightmap(a *Area) (heightmap, int) {
	h := make(heightmap, 256)
	highestY, c := a.r[0], a.c[0]

	for index := range c.sub {
		if c.sub[index].Empty() {
			continue
		}
		highestY = int(c.subY(int16(index))) + 15
	}
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			for y := highestY; y >= c.r[0]; y-- {
				if a.highest(cube.Pos{int(x) + a.baseX, y, int(z) + a.baseZ}, FilteringBlocks) == 0 {
					continue
				}
				h.set(x, z, int16(y))
				break
			}
		}
	}
	return h, highestY
}

// at returns the heightmap value at a specific column in the chunk.
func (h heightmap) at(x, z uint8) int16 {
	return h[(uint16(x)<<4)|uint16(z)]
}

// set changes the heightmap value at a specific column in the chunk.
func (h heightmap) set(x, z uint8, val int16) {
	h[(uint16(x)<<4)|uint16(z)] = val
}

// highestNeighbour returns the heightmap value of the highest directly neighbouring column of the x and z values
// passed.
func (h heightmap) highestNeighbour(x, z uint8) int16 {
	highest := int16(math.MinInt16)
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

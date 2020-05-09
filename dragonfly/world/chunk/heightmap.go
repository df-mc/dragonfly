package chunk

// heightmap represents the heightmap of a chunk. It holds the y value of all the highest blocks in the chunk
// that diffuse or obstruct light.
type heightmap []uint8

// calculateHeightmap calculates the heightmap of the chunk passed and returns it.
func calculateHeightmap(c *Chunk) heightmap {
	h := make(heightmap, 256)

	highestY := 0
	for y := range c.sub {
		if c.sub[y] != nil {
			highestY = (y << 4) + 15
		}
	}
	if highestY == 0 {
		// No sub chunks present at all.
		return h
	}

	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			for y := highestY; y >= 0; y-- {
				yVal := uint8(y)
				localYVal := uint8(y) & 0xf
				sub := subByY(yVal, c)
				if opaqueBlockPresent(sub, x, localYVal, z) {
					h.set(x, z, yVal)
					break
				} else if filterLevel(sub, x, localYVal, z) > 0 {
					h.set(x, z, yVal)
					break
				}
			}
		}
	}
	return h
}

// at returns the heightmap value at a specific column in the chunk.
func (h heightmap) at(x, z uint8) uint8 {
	return h[(uint16(x)<<4)|uint16(z)]
}

// set sets the heightmap value at a specific column in the chunk.
func (h heightmap) set(x, z, val uint8) {
	h[(uint16(x)<<4)|uint16(z)] = val
}

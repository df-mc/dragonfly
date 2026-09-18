package chunk

import "github.com/sandertv/gophertunnel/minecraft/protocol"

// SubChunkHeightMaps prepares a Chunk's surface once for multiple sub-chunk
// entries. Recreate it after editing the Chunk.
type SubChunkHeightMaps struct {
	c               *Chunk
	heights         HeightMap
	lowest, highest int16
}

// NewSubChunkHeightMaps prepares the surface summary for c.
func NewSubChunkHeightMaps(c *Chunk) SubChunkHeightMaps {
	heights := c.HeightMap()
	lowest := c.SubIndex(heights[0])
	highest := lowest
	for _, y := range heights[1:] {
		index := c.SubIndex(y)
		lowest, highest = min(lowest, index), max(highest, index)
	}
	return SubChunkHeightMaps{c: c, heights: heights, lowest: lowest, highest: highest}
}

// At returns the height-map type and optional per-column data for index.
func (m SubChunkHeightMaps) At(index int16) (byte, []int8) {
	switch {
	case index < m.lowest:
		return protocol.HeightMapDataTooHigh, nil
	case index > m.highest:
		return protocol.HeightMapDataTooLow, nil
	}
	heights := make([]int8, 256)
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			y := m.heights.At(x, z)
			columnIndex := m.c.SubIndex(y)
			i := uint16(z)<<4 | uint16(x)
			switch {
			case columnIndex > index:
				heights[i] = 16
			case columnIndex < index:
				heights[i] = -1
			default:
				heights[i] = int8(y - m.c.SubY(columnIndex))
			}
		}
	}
	return protocol.HeightMapDataHasData, heights
}

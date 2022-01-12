package chunk

import (
	"bytes"
	"github.com/df-mc/dragonfly/server/block/cube"
	"math"
)

// Area represents a square area of N*N chunks. It is used for light calculation specifically.
type Area struct {
	baseX, baseZ int
	c            []*Chunk
	w            int
	r            cube.Range
}

// NewArea creates an Area with the corner of the Area at baseX and baseY. The length of the Chunk slice must be a
// square of a number, so 1, 4, 9 etc.
func NewArea(c []*Chunk, baseX, baseY int) *Area {
	w := int(math.Sqrt(float64(len(c))))
	if len(c) != w*w {
		panic("area must have a square chunk area")
	}
	return &Area{c: c, w: w, baseX: baseX << 4, baseZ: baseY << 4, r: c[0].r}
}

// light returns the light at a cube.Pos with the light type l.
func (a *Area) light(pos cube.Pos, l light) uint8 {
	return l.light(a.sub(pos), uint8(pos[0]&0xf), uint8(pos[1]&0xf), uint8(pos[2]&0xf))
}

// light sets the light at a cube.Pos with the light type l.
func (a *Area) setLight(pos cube.Pos, l light, v uint8) {
	l.setLight(a.sub(pos), uint8(pos[0]&0xf), uint8(pos[1]&0xf), uint8(pos[2]&0xf), v)
}

// neighbours returns all neighbour lightNode of the one passed. If one of these nodes would otherwise fall outside the
// Area, it is not returned.
func (a *Area) neighbours(n lightNode) []lightNode {
	nodes := make([]lightNode, 0, 6)
	for _, f := range cube.Faces() {
		nn := lightNode{pos: n.pos.Side(f), lt: n.lt}
		if nn.pos[1] <= a.r.Max() && nn.pos[1] >= a.r.Min() && nn.pos[0] >= a.baseX && nn.pos[2] >= a.baseZ && nn.pos[0] < a.baseX+a.w*16 && nn.pos[2] < a.baseZ+a.w*16 {
			nodes = append(nodes, nn)
		}
	}
	return nodes
}

// iterSubChunks iterates over all blocks of the Area on a per-SubChunk basis. A filter function may be passed to
// specify if a SubChunk should be iterated over. If it returns false, it will not be iterated over.
func (a *Area) iterSubChunks(filter func(sub *SubChunk) bool, f func(pos cube.Pos)) {
	for cx := 0; cx < a.w; cx++ {
		for cz := 0; cz < a.w; cz++ {
			baseX, baseZ, c := a.baseX+(cx<<4), a.baseZ+(cz<<4), a.c[a.chunkIndex(cx, cz)]

			for index, sub := range c.sub {
				if !filter(sub) {
					continue
				}
				baseY := int(c.subY(int16(index)))
				a.iterSubChunk(func(x, y, z int) {
					f(cube.Pos{x + baseX, y + baseY, z + baseZ})
				})
			}
		}
	}
}

// iterEdges iterates over all chunk edges within the Area and calls the function f with the cube.Pos at either side
// of the edge.
func (a *Area) iterEdges(f func(a, b cube.Pos)) {
	width := a.w * 16
	for cx := 1; cx < a.w; cx++ {
		x := a.baseX + (cx << 4)
		for z := a.baseZ; z < a.baseZ+width; z++ {
			for y := a.r[0]; y < a.r[1]; y++ {
				f(cube.Pos{x, y, z}, cube.Pos{x - 1, y, z})
			}
		}
	}
	for cz := 1; cz < a.w; cz++ {
		z := a.baseZ + (cz << 4)
		for x := a.baseX; x < a.baseX+width; x++ {
			for y := a.r[0]; y < a.r[1]; y++ {
				f(cube.Pos{x, y, z}, cube.Pos{x, y, z - 1})
			}
		}
	}
}

// iterHeightmap iterates over the heightmap of the Area and calls the function f with the heightmap value, the
// heightmap value of the highest neighbour and the Y value of the highest non-empty SubChunk.
func (a *Area) iterHeightmap(f func(x, z int, height, highestNeighbour, highestY int)) {
	m, highestY := calculateHeightmap(a)
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			f(int(x)+a.baseX, int(z)+a.baseZ, int(m.at(x, z)), int(m.highestNeighbour(x, z)), highestY)
		}
	}
}

// iterSubChunk iterates over the coordinates of a SubChunk (0-15 on all axes) and calls the function f for each of
// those coordinates.
func (a *Area) iterSubChunk(f func(x, y, z int)) {
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			for z := 0; z < 16; z++ {
				f(x, y, z)
			}
		}
	}
}

// highest looks up through the blocks at first and second layer at the cube.Pos passed and runs their runtime IDs
// through the slice m passed, finding the highest value in this slice between those runtime IDs and returning it.
func (a *Area) highest(pos cube.Pos, m []uint8) uint8 {
	x, y, z, sub := uint8(pos[0]&0xf), uint8(pos[1]&0xf), uint8(pos[2]&0xf), a.sub(pos)
	storages, l := sub.storages, len(sub.storages)

	switch l {
	case 0:
		return 0
	case 1:
		return m[storages[0].At(x, y, z)]
	default:
		level := m[storages[0].At(x, y, z)]
		if v := m[storages[1].At(x, y, z)]; v > level {
			return v
		}
		return level
	}
}

var (
	fullLight    = bytes.Repeat([]byte{0xff}, 2048)
	fullLightPtr = &fullLight[0]
	noLight      = make([]uint8, 2048)
	noLightPtr   = &noLight[0]
)

// initialiseLightSlices initialises all light slices in the sub chunks of all chunks either with full light if there is
// no sub chunk with any blocks above it, or with empty light if there is. The sub chunks with empty light are then
// ready to be properly calculated.
func (a *Area) initialiseLightSlices() {
	for _, c := range a.c {
		index := len(c.sub) - 1
		for index >= 0 {
			sub := c.sub[index]
			if !sub.Empty() {
				// We've hit the topmost empty SubChunk.
				break
			}
			sub.skyLight = fullLight
			sub.blockLight = noLight
			index--
		}
		for index >= 0 {
			// Fill up the rest of the sub chunks with empty light. We will do light calculation for these sub chunks
			// later on.
			c.sub[index].skyLight = noLight
			c.sub[index].blockLight = noLight
			index--
		}
	}
}

// sub returns the SubChunk corresponding to a cube.Pos.
func (a *Area) sub(pos cube.Pos) *SubChunk {
	return a.chunk(pos).subChunk(int16(pos[1]))
}

// chunk returns the Chunk corresponding to a cube.Pos.
func (a *Area) chunk(pos cube.Pos) *Chunk {
	x, z := pos[0]-a.baseX, pos[2]-a.baseZ
	return a.c[a.chunkIndex(x, z)]
}

// chunkIndex finds the index in the c slices of an Area for a Chunk at a specific x and z.
func (a *Area) chunkIndex(x, z int) int {
	// (-1 -1), (0 -1), (1 -1)
	// (-1 0),  (0 0),  (1 0)
	// (-1 1),  (0 1),  (1 1)
	return (x >> 4) | ((z >> 4) * a.w)
}

package chunk

import (
	"bytes"
	"container/list"
	"github.com/df-mc/dragonfly/server/block/cube"
	"math"
)

// lightArea represents a square area of N*N chunks. It is used for light calculation specifically.
type lightArea struct {
	baseX, baseZ int
	c            []*Chunk
	w            int
	r            cube.Range
}

// LightArea creates a lightArea with the lower corner of the lightArea at baseX and baseY. The length of the Chunk
// slice must be a square of a number, so 1, 4, 9 etc.
func LightArea(c []*Chunk, baseX, baseY int) *lightArea {
	w := int(math.Sqrt(float64(len(c))))
	if len(c) != w*w {
		panic("area must have a square chunk area")
	}
	return &lightArea{c: c, w: w, baseX: baseX << 4, baseZ: baseY << 4, r: c[0].r}
}

// Fill executes the light 'filling' stage, where the lightArea is filled with light coming only from the
// individual chunks within the lightArea itself, without light crossing chunk borders.
func (a *lightArea) Fill() {
	a.initialiseLightSlices()
	queue := list.New()
	a.insertBlockLightNodes(queue)
	a.insertSkyLightNodes(queue)

	for queue.Len() != 0 {
		a.propagate(queue)
	}
}

// Spread executes the light 'spreading' stage, where the lightArea has light spread from every Chunk into the
// neighbouring chunks. The neighbouring chunks must have passed the light 'filling' stage before this
// function is called for an lightArea that includes them.
func (a *lightArea) Spread() {
	queue := list.New()
	a.insertLightSpreadingNodes(queue, BlockLight)
	a.insertLightSpreadingNodes(queue, SkyLight)

	for queue.Len() != 0 {
		a.propagate(queue)
	}
}

// light returns the light at a cube.Pos with the light type l.
func (a *lightArea) light(pos cube.Pos, l light) uint8 {
	return l.light(a.sub(pos), uint8(pos[0]&0xf), uint8(pos[1]&0xf), uint8(pos[2]&0xf))
}

// light sets the light at a cube.Pos with the light type l.
func (a *lightArea) setLight(pos cube.Pos, l light, v uint8) {
	l.setLight(a.sub(pos), uint8(pos[0]&0xf), uint8(pos[1]&0xf), uint8(pos[2]&0xf), v)
}

// neighbours returns all neighbour lightNode of the one passed. If one of these nodes would otherwise fall outside the
// lightArea, it is not returned.
func (a *lightArea) neighbours(n lightNode) []lightNode {
	nodes := make([]lightNode, 0, 6)
	for _, f := range cube.Faces() {
		nn := lightNode{pos: n.pos.Side(f), lt: n.lt}
		if nn.pos[1] <= a.r.Max() && nn.pos[1] >= a.r.Min() && nn.pos[0] >= a.baseX && nn.pos[2] >= a.baseZ && nn.pos[0] < a.baseX+a.w*16 && nn.pos[2] < a.baseZ+a.w*16 {
			nodes = append(nodes, nn)
		}
	}
	return nodes
}

// iterSubChunks iterates over all blocks of the lightArea on a per-SubChunk basis. A filter function may be passed to
// specify if a SubChunk should be iterated over. If it returns false, it will not be iterated over.
func (a *lightArea) iterSubChunks(filter func(sub *SubChunk) bool, f func(pos cube.Pos)) {
	for cx := 0; cx < a.w; cx++ {
		for cz := 0; cz < a.w; cz++ {
			baseX, baseZ, c := a.baseX+(cx<<4), a.baseZ+(cz<<4), a.c[a.chunkIndex(cx, cz)]

			for index, sub := range c.sub {
				if !filter(sub) {
					continue
				}
				baseY := int(c.SubY(int16(index)))
				a.iterSubChunk(func(x, y, z int) {
					f(cube.Pos{x + baseX, y + baseY, z + baseZ})
				})
			}
		}
	}
}

// iterEdges iterates over all chunk edges within the lightArea and calls the function f with the cube.Pos at either
// side of the edge.
func (a *lightArea) iterEdges(filter func(a, b *SubChunk) bool, f func(a, b cube.Pos)) {
	minY, maxY := a.r[0]>>4, a.r[1]>>4
	// First iterate over chunk X, Y and Z, so we can filter out a complete 16x16 sheet of blocks if the
	// filter function returns false.
	for cu := 1; cu < a.w; cu++ {
		u := cu << 4
		for cv := 0; cv < a.w; cv++ {
			v := cv << 4
			for cy := minY; cy < maxY; cy++ {
				baseY := cy << 4

				xa, za := cube.Pos{a.baseX + u, baseY, a.baseZ + v}, cube.Pos{a.baseX + v, baseY, a.baseZ + u}
				xb, zb := xa.Side(cube.FaceWest), za.Side(cube.FaceNorth)

				addX, addZ := filter(a.sub(xa), a.sub(xb)), filter(a.sub(za), a.sub(zb))
				if !addX && !addZ {
					continue
				}
				// The order of these loops allows us to take care of block spreading over both the X and Z axis by
				// just swapping around the axes.
				for addV := 0; addV < 16; addV++ {
					for y := 0; y < 16; y++ {
						// Finally, iterate over the 16x16 sheet and actually do the per-block checks.
						if addX {
							f(xa.Add(cube.Pos{0, y, addV}), xb.Add(cube.Pos{0, y, addV}))
						}
						if addZ {
							f(za.Add(cube.Pos{addV, y}), zb.Add(cube.Pos{addV, y}))
						}
					}
				}
			}
		}
	}
}

// iterHeightmap iterates over the height map of the lightArea and calls the function f with the height map value, the
// height map value of the highest neighbour and the Y value of the highest non-empty SubChunk.
func (a *lightArea) iterHeightmap(f func(x, z int, height, highestNeighbour, highestY int)) {
	m, highestY := a.c[0].HeightMap(), a.c[0].Range().Min()
	for index := range a.c[0].sub {
		if a.c[0].sub[index].Empty() {
			continue
		}
		highestY = int(a.c[0].SubY(int16(index))) + 15
	}
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			f(int(x)+a.baseX, int(z)+a.baseZ, int(m.At(x, z)), int(m.HighestNeighbour(x, z)), highestY)
		}
	}
}

// iterSubChunk iterates over the coordinates of a SubChunk (0-15 on all axes) and calls the function f for each of
// those coordinates.
func (a *lightArea) iterSubChunk(f func(x, y, z int)) {
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
func (a *lightArea) highest(pos cube.Pos, m []uint8) uint8 {
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
func (a *lightArea) initialiseLightSlices() {
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
func (a *lightArea) sub(pos cube.Pos) *SubChunk {
	return a.chunk(pos).SubChunk(int16(pos[1]))
}

// chunk returns the Chunk corresponding to a cube.Pos.
func (a *lightArea) chunk(pos cube.Pos) *Chunk {
	x, z := pos[0]-a.baseX, pos[2]-a.baseZ
	return a.c[a.chunkIndex(x>>4, z>>4)]
}

// chunkIndex finds the index in the chunk slice of an lightArea for a Chunk at a specific x and z.
func (a *lightArea) chunkIndex(x, z int) int {
	return x + (z * a.w)
}

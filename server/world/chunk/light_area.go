package chunk

import (
	"bytes"
	"github.com/df-mc/dragonfly/server/block/cube"
)

type Area struct {
	baseX, baseZ int
	c            []*Chunk
	w            int
}

func NewArea(c []*Chunk, w int, baseX, baseY int) *Area {
	if len(c) != w*w {
		panic("chunk count must be equal to w*w")
	}
	return &Area{c: c, w: w, baseX: baseX << 4, baseZ: baseY << 4}
}

func (a *Area) Light(pos cube.Pos, l light) uint8 {
	return l.light(a.sub(pos), uint8(pos[0]&0xf), uint8(pos[1]&0xf), uint8(pos[2]&0xf))
}

func (a *Area) SetLight(pos cube.Pos, l light, v uint8) {
	l.setLight(a.sub(pos), uint8(pos[0]&0xf), uint8(pos[1]&0xf), uint8(pos[2]&0xf), v)
}

func (a *Area) Neighbours(n lightNode) []lightNode {
	nodes := make([]lightNode, 0, 6)
	for _, f := range cube.Faces() {
		nn := lightNode{pos: n.pos.Side(f), lt: n.lt}
		if nn.pos[1] <= cube.MaxY && nn.pos[1] >= cube.MinY && nn.pos[0] >= a.baseX && nn.pos[2] >= a.baseZ && nn.pos[0] < a.baseX+a.w*16 && nn.pos[2] < a.baseZ+a.w*16 {
			nodes = append(nodes, nn)
		}
	}
	return nodes
}

func (a *Area) horizontalNeighbours(pos cube.Pos) []cube.Pos {
	positions := make([]cube.Pos, 0, 4)
	for _, f := range cube.HorizontalFaces() {
		p := pos.Side(f)
		if p[1] <= cube.MaxY && p[1] >= cube.MinY && p[0] >= a.baseX && p[2] >= a.baseZ && p[0] < a.baseX+a.w*16 && p[2] < a.baseZ+a.w*16 {
			positions = append(positions, p)
		}
	}
	return positions
}

func (a *Area) IterSubChunks(filter func(sub *SubChunk) bool, f func(pos cube.Pos)) {
	for cx := 0; cx < a.w; cx++ {
		for cz := 0; cz < a.w; cz++ {
			baseX, baseZ, c := a.baseX+(cx<<4), a.baseZ+(cz<<4), a.c[a.chunkIndex(cx, cz)]

			for index, sub := range c.sub {
				if filter(sub) {
					baseY := int(subY(int16(index)))
					a.iterSubChunk(func(x, y, z int) {
						f(cube.Pos{x + baseX, y + baseY, z + baseZ})
					})
				}
			}
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

func (a *Area) Highest(pos cube.Pos, m []uint8) uint8 {
	x, y, z, sub := uint8(pos[0]&0xf), uint8(pos[1]&0xf), uint8(pos[2]&0xf), a.sub(pos)
	storages, l := sub.storages, len(sub.storages)

	var level uint8
	if l > 0 {
		if id := storages[0].At(x, y, z); id != sub.air {
			level = m[id]
		}
		if l > 1 {
			if id := storages[1].At(x, y, z); id != sub.air {
				if v := m[id]; v > level {
					level = v
				}
			}
		}
	}
	return level
}

var (
	fullLight    = bytes.Repeat([]byte{0xff}, 2048)
	fullLightPtr = &fullLight[0]
	noLight      = make([]uint8, 2048)
	noLightPtr   = &noLight[0]
)

func (a *Area) initialiseLightSlices() {
	for _, c := range a.c {
		index := len(c.sub) - 1
		for index >= 0 {
			if sub := c.sub[index]; sub.Empty() {
				sub.skyLight = fullLight
				sub.blockLight = noLight
				index--
				continue
			}
			// We've hit the topmost empty SubChunk.
			break
		}
		for index >= 0 {
			c.sub[index].skyLight = noLight
			c.sub[index].blockLight = noLight
			index--
		}
	}
}

func (a *Area) sub(pos cube.Pos) *SubChunk {
	return a.chunk(pos).sub[subIndex(int16(pos[1]))]
}

func (a *Area) chunk(pos cube.Pos) *Chunk {
	x, z := pos[0]-a.baseX, pos[2]-a.baseZ
	return a.c[a.chunkIndex(x, z)]
}

func (a *Area) chunkIndex(x, z int) int {
	return (x >> 4) | ((z >> 4) * a.w)
}

package chunk

import "github.com/df-mc/dragonfly/server/block/cube"

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
		node := lightNode{pos: n.pos.Side(f), lt: n.lt}
		if n.pos[1] > cube.MaxY || n.pos[1] < cube.MinY || n.pos[0] < a.baseX || n.pos[2] < a.baseZ || n.pos[0] >= a.baseX+a.w*16 || n.pos[2] >= a.baseZ+a.w*16 {
			continue
		}
		nodes = append(nodes, node)
	}
	return nodes
}

func (a *Area) Highest(pos cube.Pos, m []uint8) uint8 {
	sub := a.sub(pos)
	x, y, z := uint8(pos[0]&0xf), uint8(pos[1]&0xf), uint8(pos[2]&0xf)
	var level uint8

	storages, l := sub.storages, len(sub.storages)
	if l > 0 {
		if id := storages[0].RuntimeID(x, y, z); id != sub.air {
			level = m[id]
		}
		if l > 1 {
			if id := storages[0].RuntimeID(x, y, z); id != sub.air {
				if v := m[id]; v > level {
					level = v
				}
			}
		}
	}
	return level
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

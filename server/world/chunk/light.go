package chunk

import (
	"container/list"
	"github.com/df-mc/dragonfly/server/block/cube"
)

// FillLight executes the light 'filling' stage, where the chunk is filled with light coming only from the
// chunk itself, without light crossing chunk borders.
func FillLight(a *Area) {
	a.initialiseLightSlices()
	queue := list.New()
	insertBlockLightNodes(queue, a)
	insertSkyLightNodes(queue, a)

	for queue.Len() != 0 {
		propagate(queue, a)
	}
}

// SpreadLight executes the light 'spreading' stage, where the chunk has its light spread into the
// neighbouring chunks. The neighbouring chunks must have passed the light 'filling' stage before this
// function is called for a chunk.
func SpreadLight(a *Area) {
	queue := list.New()
	insertLightSpreadingNodes(queue, a, BlockLight)
	insertLightSpreadingNodes(queue, a, SkyLight)

	for queue.Len() != 0 {
		propagate(queue, a)
	}
}

// insertBlockLightNodes iterates over the chunk and looks for blocks that have a light level of at least 1.
// If one is found, a node is added for it to the node queue.
func insertBlockLightNodes(queue *list.List, a *Area) {
	a.IterSubChunks(anyLightBlocks, func(pos cube.Pos) {
		if level := a.Highest(pos, LightBlocks); level > 0 {
			queue.PushBack(node(pos, level, BlockLight))
		}
	})
}

// anyLightBlocks checks if there are any blocks in the SubChunk passed that emit light.
func anyLightBlocks(sub *SubChunk) bool {
	for _, layer := range sub.storages {
		for _, id := range layer.palette.values {
			if LightBlocks[id] != 0 {
				return true
			}
		}
	}
	return false
}

// anyLight returns a function that can be used to check if a SubChunk has any light of the type passed.
func anyLight(lt light) func(sub *SubChunk) bool {
	if lt == SkyLight {
		return func(sub *SubChunk) bool {
			return &sub.skyLight[0] == noLightPtr
		}
	}
	return func(sub *SubChunk) bool {
		return &sub.blockLight[0] == noLightPtr
	}
}

// insertSkyLightNodes iterates over the chunk and inserts a light node anywhere at the highest block in the
// chunk. In addition, any skylight above those nodes will be set to 15.
func insertSkyLightNodes(queue *list.List, a *Area) {
	for cx := 0; cx < a.w; cx++ {
		for cz := 0; cz < a.w; cz++ {
			c := a.c[a.chunkIndex(cx, cz)]
			baseX, baseZ := a.baseX+(cx<<4), a.baseZ+(cz<<4)

			m := calculateHeightmap(c)
			highestY := c.r[0]
			for index := range c.sub {
				if c.sub[index] != nil {
					highestY = int(c.subY(int16(index)) + 15)
				}
			}
			for x := uint8(0); x < 16; x++ {
				for z := uint8(0); z < 16; z++ {
					current := int(m.at(x, z))
					highestNeighbour := int(m.highestNeighbour(x, z))

					// If we hit a block like water or leaves (something that diffuses but does not block light), we
					// need a node above this block regardless of the neighbours.
					pos := cube.Pos{baseX + int(x), current, baseZ + int(z)}
					if level := a.Highest(pos, FilteringBlocks); level < 14 && level > 0 {
						queue.PushBack(node(pos.Add(cube.Pos{0, 1}), 15, SkyLight))
						continue
					}
					for y := current; y < highestY; y++ {
						// We can do a bit of an optimisation here: We don't need to insert nodes if the neighbours are
						// lower than the current one, on the same index level, or one level higher, because light in
						// this column can't spread below that anyway.
						pos[1] = y + 1
						if y < highestNeighbour-1 {
							queue.PushBack(node(pos, 15, SkyLight))
							continue
						}
						// Fill the rest of the column with full skylight.
						a.SetLight(pos, SkyLight, 15)
					}
				}
			}
		}
	}
}

// insertLightSpreadingNodes inserts light nodes into the node queue passed which, when propagated, will
// spread into the neighbouring chunks.
func insertLightSpreadingNodes(queue *list.List, a *Area, lt light) {
	a.IterSubChunks(anyLight(lt), func(pos cube.Pos) {
		if lx, lz := uint8(pos[0]&0xf), uint8(pos[2]&0xf); lx != 0 && lx != 15 && lz != 0 && lz != 15 {
			return
		}
		if l := a.Light(pos, lt); l > 1 {
			for _, n := range a.horizontalNeighbours(pos) {
				if (n[0]>>4 != pos[0]>>4 || n[2]>>4 != pos[2]>>4) && a.Light(n, lt) < l {
					queue.PushBack(node(pos, l, lt))
					break
				}
			}
		}
	})
}

// propagate spreads the next light node in the node queue passed through the Area a. propagate adds the neighbours
// of the node to the queue for as long as it is able to spread.
func propagate(queue *list.List, a *Area) {
	n := queue.Remove(queue.Front()).(lightNode)
	if a.Light(n.pos, n.lt) > n.level {
		return
	}
	a.SetLight(n.pos, n.lt, n.level)

	for _, neighbour := range a.Neighbours(n) {
		filter := a.Highest(neighbour.pos, FilteringBlocks) + 1
		next := n.level - filter

		if n.level > filter && a.Light(neighbour.pos, n.lt) < next {
			neighbour.level = next
			queue.PushBack(neighbour)
		}
	}
}

// lightNode is a node pushed to the queue which is used to propagate light.
type lightNode struct {
	pos   cube.Pos
	lt    light
	level uint8
}

// node creates a new lightNode using the position, level and light type passed.
func node(pos cube.Pos, level uint8, lt light) lightNode {
	return lightNode{pos: pos, level: level, lt: lt}
}

// filterLevel checks for the block with the Highest filter level in the sub chunk at a specific position,
// returning 15 if there is a block, but if it is not present in the FilteringBlocks map.
func filterLevel(sub *SubChunk, x, y, z uint8) uint8 {
	storages := sub.storages
	// We offer several fast ways out to get a little more performance out of this.
	switch len(storages) {
	case 0:
		return 0
	case 1:
		id := storages[0].At(x, y, z)
		if id == sub.air {
			return 0
		}
		return FilteringBlocks[id]
	case 2:
		var highest uint8

		id := storages[0].At(x, y, z)
		if id != sub.air {
			highest = FilteringBlocks[id]
		}

		id = storages[1].At(x, y, z)
		if id != sub.air {
			if v := FilteringBlocks[id]; v > highest {
				highest = v
			}
		}
		return highest
	}
	var highest uint8
	for i := range storages {
		id := storages[i].At(x, y, z)
		if id != sub.air {
			if l := FilteringBlocks[id]; l > highest {
				highest = l
			}
		}
	}
	return highest
}

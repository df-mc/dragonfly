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

// insertSkyLightNodes iterates over the chunk and inserts a light node anywhere at the highest block in the
// chunk. In addition, any skylight above those nodes will be set to 15.
func insertSkyLightNodes(queue *list.List, a *Area) {
	for cx := 0; cx < a.w; cx++ {
		for cz := 0; cz < a.w; cz++ {
			baseX, baseZ := a.baseX+(cx<<4), a.baseZ+(cz<<4)

			m, highestY := calculateHeightmap(a)
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
	a.IterEdges(func(pa, pb cube.Pos) {
		la, lb := a.Light(pa, lt), a.Light(pb, lt)
		if res := la - a.Highest(pb, FilteringBlocks); res > lb {
			queue.PushBack(node(pb, res, lt))
		} else if res = lb - a.Highest(pa, FilteringBlocks); res > la {
			queue.PushBack(node(pa, res, lt))
		}
	})
}

// propagate spreads the next light node in the node queue passed through the Area a. propagate adds the neighbours
// of the node to the queue for as long as it is able to spread.
func propagate(queue *list.List, a *Area) {
	n := queue.Remove(queue.Front()).(lightNode)
	if a.Light(n.pos, n.lt) >= n.level {
		return
	}
	a.SetLight(n.pos, n.lt, n.level)

	for _, neighbour := range a.Neighbours(n) {
		filter := a.Highest(neighbour.pos, FilteringBlocks) + 1
		if n.level > filter && a.Light(neighbour.pos, n.lt) < n.level-filter {
			neighbour.level = n.level - filter
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

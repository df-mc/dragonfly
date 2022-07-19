package chunk

import (
	"container/list"
	"github.com/df-mc/dragonfly/server/block/cube"
)

// insertBlockLightNodes iterates over the chunk and looks for blocks that have a light level of at least 1.
// If one is found, a node is added for it to the node queue.
func (a *lightArea) insertBlockLightNodes(queue *list.List) {
	a.iterSubChunks(anyLightBlocks, func(pos cube.Pos) {
		if level := a.highest(pos, LightBlocks); level > 0 {
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
func (a *lightArea) insertSkyLightNodes(queue *list.List) {
	a.iterHeightmap(func(x, z int, height, highestNeighbour, highestY int) {
		// If we hit a block like water or leaves (something that diffuses but does not block light), we
		// need a node above this block regardless of the neighbours.
		pos := cube.Pos{x, height, z}
		if level := a.highest(pos, FilteringBlocks); level != 15 && level != 0 {
			queue.PushBack(node(pos.Side(cube.FaceUp), 15, SkyLight))
			pos[1]++
		}
		for y := pos[1]; y < highestY; y++ {
			// We can do a bit of an optimisation here: We don't need to insert nodes if the neighbours are
			// lower than the current one, on the same Y level, or one level higher, because light in
			// this column can't spread below that anyway.
			if pos[1]++; pos[1] < highestNeighbour {
				queue.PushBack(node(pos, 15, SkyLight))
				continue
			}
			// Fill the rest with full skylight.
			a.setLight(pos, SkyLight, 15)
		}
	})
}

// insertLightSpreadingNodes inserts light nodes into the node queue passed which, when propagated, will
// spread into the neighbouring chunks.
func (a *lightArea) insertLightSpreadingNodes(queue *list.List, lt light) {
	a.iterEdges(a.nodesNeeded(lt), func(pa, pb cube.Pos) {
		la, lb := a.light(pa, lt), a.light(pb, lt)
		if la == lb || la-1 == lb || lb-1 == la {
			// No chance for this to spread. Don't check for the highest filtering blocks on the side.
			return
		}
		if filter := a.highest(pb, FilteringBlocks) + 1; la > filter && la-filter > lb {
			queue.PushBack(node(pb, la-filter, lt))
		} else if filter = a.highest(pa, FilteringBlocks) + 1; lb > filter && lb-filter > la {
			queue.PushBack(node(pa, lb-filter, lt))
		}
	})
}

// nodesNeeded checks if any light nodes of a specific light type are needed between two neighbouring SubChunks when
// spreading light between them.
func (a *lightArea) nodesNeeded(lt light) func(sa, sb *SubChunk) bool {
	if lt == SkyLight {
		return func(sa, sb *SubChunk) bool {
			return &sa.skyLight[0] != &sb.skyLight[0]
		}
	}
	return func(sa, sb *SubChunk) bool {
		// Don't add nodes if both sub chunks are either both fully filled with light or have no light at all.
		return &sa.blockLight[0] != &sb.blockLight[0]
	}
}

// propagate spreads the next light node in the node queue passed through the lightArea a. propagate adds the neighbours
// of the node to the queue for as long as it is able to spread.
func (a *lightArea) propagate(queue *list.List) {
	n := queue.Remove(queue.Front()).(lightNode)
	if a.light(n.pos, n.lt) >= n.level {
		return
	}
	a.setLight(n.pos, n.lt, n.level)

	for _, neighbour := range a.neighbours(n) {
		filter := a.highest(neighbour.pos, FilteringBlocks) + 1
		if n.level > filter && a.light(neighbour.pos, n.lt) < n.level-filter {
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

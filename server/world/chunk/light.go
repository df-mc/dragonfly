package chunk

import (
	"container/list"
	"github.com/df-mc/dragonfly/server/block/cube"
)

// LightBlocks is a list of block light levels (0-15) indexed by block runtime IDs. The map is used to do a
// fast lookup of block light.
var LightBlocks = make([]uint8, 0, 7000)

// FilteringBlocks is a map for checking if a block runtime ID filters light, and if so, how many levels.
// Light is able to propagate through these blocks, but will have its level reduced.
var FilteringBlocks = make([]uint8, 0, 7000)

// lightNode is a node pushed to the queue which is used to propagate light.
type lightNode struct {
	x, z  int8
	y     int16
	level uint8
	first bool
}

// neighbours returns all neighbouring nodes of the current one.
func (n lightNode) neighbours() []lightNode {
	neighbours := make([]lightNode, 6)
	neighbours[0] = lightNode{x: n.x - 1, y: n.y, z: n.z}
	neighbours[1] = lightNode{x: n.x + 1, y: n.y, z: n.z}
	neighbours[2] = lightNode{x: n.x, y: n.y, z: n.z - 1}
	neighbours[3] = lightNode{x: n.x, y: n.y, z: n.z + 1}

	if n.y == cube.MaxY {
		neighbours[4] = lightNode{x: n.x, y: n.y - 1, z: n.z}
		return neighbours[:5]
	} else if n.y == cube.MinY {
		neighbours[4] = lightNode{x: n.x, y: n.y + 1, z: n.z}
		return neighbours[:5]
	}
	neighbours[4] = lightNode{x: n.x, y: n.y + 1, z: n.z}
	neighbours[5] = lightNode{x: n.x, y: n.y - 1, z: n.z}

	return neighbours
}

// FillLight executes the light 'filling' stage, where the chunk is filled with light coming only from the
// chunk itself, without light crossing chunk borders.
func FillLight(c *Chunk) {
	removeEmptySubChunks(c)
	fillBlockLight(c)
	fillSkyLight(c)
}

// SpreadLight executes the light 'spreading' stage, where the chunk has its light spread into the
// neighbouring chunks. The neighbouring chunks must have passed the light 'filling' stage before this
// function is called for a chunk.
func SpreadLight(c *Chunk, neighbours []*Chunk) {
	spreadBlockLight(c, neighbours)
	spreadSkyLight(c, neighbours)

	// Spreading light might create new sub chunks, but we don't want those as sky light might not be
	// initially spread there.
	removeEmptySubChunks(c)
	for i := range neighbours {
		removeEmptySubChunks(neighbours[i])
	}
}

// removeEmptySubChunks removes any empty sub chunks from the top of the chunk passed.
func removeEmptySubChunks(c *Chunk) {
	for index, sub := range c.sub {
		if sub == nil {
			continue
		}
		if len(sub.storages) == 0 {
			c.sub[index] = nil
		} else if len(sub.storages) == 1 && len(sub.storages[0].palette.blockRuntimeIDs) == 1 && sub.storages[0].palette.blockRuntimeIDs[0] == c.air {
			// Sub chunk with only air in it.
			c.sub[index] = nil
		} else {
			// We found a sub chunk that has blocks, so break out.
			break
		}
	}
}

// spreadSkyLight spreads the sky light from the current chunk into the chunks around it. The neighbours are
// in (-1, -1), (-1, 0), (-1, 1), (0, -1), (0, 1), (1, -1), (1, 0), (1, 1) order, with a total length of
// 8 chunks (around the centre chunk).
func spreadSkyLight(c *Chunk, neighbourChunks []*Chunk) {
	queue := list.New()
	insertSkyLightSpreadingNodes(queue, c, neighbourChunks)
	for queue.Len() != 0 {
		spreadPropagate(queue, c, neighbourChunks, true)
	}
}

// spreadBlockLight spreads the block light from the current chunk into the chunks around it. The neighbours
// are in (-1, -1), (-1, 0), (-1, 1), (0, -1), (0, 1), (1, -1), (1, 0), (1, 1) order, with a total length of
// 8 chunks (around the centre chunk).
func spreadBlockLight(c *Chunk, neighbourChunks []*Chunk) {
	queue := list.New()
	insertBlockLightSpreadingNodes(queue, c, neighbourChunks)
	for queue.Len() != 0 {
		spreadPropagate(queue, c, neighbourChunks, false)
	}
}

// fillSkyLight fills the chunk passed with sky light that has its source only within the bounds of the chunk
// passed.
func fillSkyLight(c *Chunk) {
	queue := list.New()
	insertSkyLightNodes(queue, c)
	for queue.Len() != 0 {
		fillPropagate(queue, c, true)
	}
}

// fillBlockLight fills the chunk passed with block light that has its source only within the bounds of the
// chunk passed.
func fillBlockLight(c *Chunk) {
	queue := list.New()
	if anyBlockLight(c) {
		insertBlockLightNodes(queue, c)
		for queue.Len() != 0 {
			fillPropagate(queue, c, false)
		}
	}
}

// anyBlockLight checks if there are any blocks in the Chunk passed that emit light.
func anyBlockLight(c *Chunk) bool {
	for _, sub := range c.sub {
		if sub == nil {
			continue
		}
		for _, layer := range sub.storages {
			for _, id := range layer.palette.blockRuntimeIDs {
				if LightBlocks[id] != 0 {
					return true
				}
			}
		}
	}
	return false
}

// insertSkyLightNodes iterates over the chunk and inserts a light node anywhere at the highest block in the
// chunk. In addition, any sky light above those nodes will be set to 15.
func insertSkyLightNodes(queue *list.List, c *Chunk) {
	m := calculateHeightmap(c)
	highestY := int16(cube.MinY)
	for index := range c.sub {
		if c.sub[index] != nil {
			highestY = subY(int16(index)) + 15
		}
	}
	for x := uint8(0); x < 16; x++ {
		for z := uint8(0); z < 16; z++ {
			current := m.at(x, z)
			highestNeighbour := current

			if x != 15 {
				if val := m.at(x+1, z); val > highestNeighbour {
					highestNeighbour = val
				}
			}
			if x != 0 {
				if val := m.at(x-1, z); val > highestNeighbour {
					highestNeighbour = val
				}
			}
			if z != 15 {
				if val := m.at(x, z+1); val > highestNeighbour {
					highestNeighbour = val
				}
			}
			if z != 0 {
				if val := m.at(x, z-1); val > highestNeighbour {
					highestNeighbour = val
				}
			}

			// We can do a bit of an optimisation here: We don't need to insert nodes if the neighbours are
			// lower than the current one, on the same index level, or one level higher, because light in this
			// column can't spread below that anyway.
			for y := current; y < highestY; y++ {
				if y == current {
					level := filterLevel(c.sub[subIndex(y)], x, uint8(y&0xf), z)
					if level < 14 && level > 0 {
						// If we hit a block like water or leaves, we need a node above this block regardless
						// of the neighbours.
						queue.PushBack(lightNode{x: int8(x), z: int8(z), y: y + 1, level: 15})
						continue
					}
				}
				if y < highestNeighbour-1 {
					queue.PushBack(lightNode{x: int8(x), z: int8(z), y: y + 1, level: 15})
					continue
				}
				// Fill the rest of the column with sky light on full strength.
				c.sub[subIndex(y+1)].setSkyLight(x, uint8((y+1)&0xf), z, 15)
			}
		}
	}
}

// insertBlockLightNodes iterates over the chunk and looks for blocks that have a light level of at least 1.
// If one is found, a node is added for it to the node queue.
func insertBlockLightNodes(queue *list.List, c *Chunk) {
	for index, sub := range c.sub {
		if sub == nil {
			continue
		}
		baseY := subY(int16(index))
		for y := uint8(0); y < 16; y++ {
			actualY := int16(y) + baseY
			for x := uint8(0); x < 16; x++ {
				for z := uint8(0); z < 16; z++ {
					if level := highestEmissionLevel(sub, x, y, z); level > 0 {
						queue.PushBack(lightNode{x: int8(x), z: int8(z), y: actualY, level: level})
					}
				}
			}
		}
	}
}

// insertSkyLightSpreadingNodes inserts light nodes into the node queue passed which, when propagated, will
// spread into the neighbouring chunks.
func insertSkyLightSpreadingNodes(queue *list.List, c *Chunk, neighbours []*Chunk) {
	for index, sub := range c.sub {
		if sub == nil {
			continue
		}
		baseY := subY(int16(index))
		for y := uint8(0); y < 16; y++ {
			totalY := int16(y) + baseY
			for x := uint8(0); x < 16; x++ {
				for z := uint8(0); z < 16; z++ {
					if z != 0 && z != 15 && x != 0 && x != 15 {
						break
					}
					l := sub.SkyLightAt(x, y, z)
					if l <= 1 {
						// The light level was either 0 or 1, meaning it cannot propagate either way.
						continue
					}
					nodeNeeded := false
					if x == 0 {
						subNeighbour := neighbours[1].sub[index]
						if subNeighbour != nil && subNeighbour.SkyLightAt(15, y, z) < l {
							nodeNeeded = true
						}
					} else if x == 15 {
						subNeighbour := neighbours[6].sub[index]
						if subNeighbour != nil && subNeighbour.SkyLightAt(0, y, z) < l {
							nodeNeeded = true
						}
					}
					if z == 0 {
						subNeighbour := neighbours[3].sub[index]
						if subNeighbour != nil && subNeighbour.SkyLightAt(x, y, 15) < l {
							nodeNeeded = true
						}
					} else if z == 15 {
						subNeighbour := neighbours[4].sub[index]
						if subNeighbour != nil && subNeighbour.SkyLightAt(x, y, 0) < l {
							nodeNeeded = true
						}
					}
					if nodeNeeded {
						queue.PushBack(lightNode{x: int8(x), y: totalY, z: int8(z), level: l, first: true})
					}
				}
			}
		}
	}
}

// insertSkyLightSpreadingNodes inserts block light nodes into the node queue passed which, when propagated,
// will spread into the neighbouring chunks.
func insertBlockLightSpreadingNodes(queue *list.List, c *Chunk, neighbours []*Chunk) {
	for index, sub := range c.sub {
		if sub == nil {
			continue
		}
		baseY := subY(int16(index))
		for y := uint8(0); y < 16; y++ {
			totalY := int16(y) + baseY
			for x := uint8(0); x < 16; x++ {
				for z := uint8(0); z < 16; z++ {
					if z != 0 && z != 15 && x != 0 && x != 15 {
						break
					}
					l := sub.blockLightAt(x, y, z)
					if l <= 1 {
						// The light level was either 0 or 1, meaning it cannot propagate either way.
						continue
					}
					nodeNeeded := false
					if x == 0 {
						subNeighbour := neighbours[1].sub[index]
						if subNeighbour != nil && subNeighbour.blockLightAt(15, y, z) < l {
							nodeNeeded = true
						}
					} else if x == 15 {
						subNeighbour := neighbours[6].sub[index]
						if subNeighbour != nil && subNeighbour.blockLightAt(0, y, z) < l {
							nodeNeeded = true
						}
					}
					if z == 0 {
						subNeighbour := neighbours[3].sub[index]
						if subNeighbour != nil && subNeighbour.blockLightAt(x, y, 15) < l {
							nodeNeeded = true
						}
					} else if z == 15 {
						subNeighbour := neighbours[4].sub[index]
						if subNeighbour != nil && subNeighbour.blockLightAt(x, y, 0) < l {
							nodeNeeded = true
						}
					}
					if nodeNeeded {
						queue.PushBack(lightNode{x: int8(x), y: totalY, z: int8(z), level: l, first: true})
					}
				}
			}
		}
	}
}

// spreadPropagate propagates a sky light node in the queue past through the chunk passed and its neighbours,
// unlike fillPropagate, which only propagates within the chunk.
func spreadPropagate(queue *list.List, c *Chunk, neighbourChunks []*Chunk, skylight bool) {
	node := queue.Remove(queue.Front()).(lightNode)

	x, y, z := uint8(node.x&0xf), node.y, uint8(node.z&0xf)
	yLocal := uint8(y & 0xf)
	sub := subByY(y, chunkByNode(node, c, neighbourChunks))

	if skylight {
		if !node.first {
			filter := filterLevel(sub, x, yLocal, z) + 1
			if filter >= node.level {
				return
			}
			node.level -= filter
			if sub.SkyLightAt(x, yLocal, z) >= node.level {
				// This neighbour already had either as high of a level as what we're updating it to, or
				// higher already, so spreading it further is pointless as that will already have been done.
				return
			}
			sub.setSkyLight(x, yLocal, z, node.level)
		}
	} else {
		if !node.first {
			filter := filterLevel(sub, x, yLocal, z) + 1
			if filter >= node.level {
				return
			}
			node.level -= filter
			if sub.blockLightAt(x, yLocal, z) >= node.level {
				// This neighbour already had either as high of a level as what we're updating it to, or
				// higher already, so spreading it further is pointless as that will already have been done.
				return
			}
			sub.setBlockLight(x, yLocal, z, node.level)
		}
	}
	for _, neighbour := range node.neighbours() {
		neighbour.level = node.level
		queue.PushBack(neighbour)
	}
}

// fillPropagate propagates a sky light node in the node queue passed within the chunk itself. It does not
// spread the light beyond the chunk.
func fillPropagate(queue *list.List, c *Chunk, skyLight bool) {
	node := queue.Remove(queue.Front()).(lightNode)

	x, y, z := uint8(node.x), node.y, uint8(node.z)
	yLocal := uint8(y & 0xf)
	sub := subByY(y, c)

	if skyLight {
		if sub.SkyLightAt(x, yLocal, z) >= node.level {
			// This neighbour already had either as high of a level as what we're updating it to, or
			// higher already, so spreading it further is pointless as that will already have been done.
			return
		}
		sub.setSkyLight(x, yLocal, z, node.level)
	} else {
		if sub.blockLightAt(x, yLocal, z) >= node.level {
			// This neighbour already had either as high of a level as what we're updating it to, or
			// higher already, so spreading it further is pointless as that will already have been done.
			return
		}
		sub.setBlockLight(x, yLocal, z, node.level)
	}

	// If the level is 1 or lower, it won't be able to propagate any further.
	if node.level > 1 {
		for _, neighbour := range node.neighbours() {
			if neighbour.x < 0 || neighbour.x > 15 || neighbour.z < 0 || neighbour.z > 15 {
				// In the fill stage, we don't propagate sky light out of the chunk.
				continue
			}
			sub := filterLevel(subByY(neighbour.y, c), uint8(neighbour.x), uint8(neighbour.y&0xf), uint8(neighbour.z)) + 1
			if sub >= node.level {
				// No light left to propagate.
				continue
			}
			neighbour.level = node.level - sub
			queue.PushBack(neighbour)
		}
	}
}

// subByY returns a sub chunk in the chunk passed by a Y value. If one doesn't yet exist, it is created.
func subByY(y int16, c *Chunk) *SubChunk {
	index := subIndex(y)
	sub := c.sub[index]

	if sub == nil {
		sub = NewSubChunk(c.air)
		c.sub[index] = sub
	}
	return sub
}

// chunkByNode selects a chunk (either the centre or one of the neighbours) depending on the position of the
// node passed.
func chunkByNode(node lightNode, centre *Chunk, neighbours []*Chunk) *Chunk {
	switch {
	case node.x < 0 && node.z < 0:
		return neighbours[0]
	case node.x < 0 && node.z >= 0 && node.z <= 15:
		return neighbours[1]
	case node.x < 0 && node.z >= 16:
		return neighbours[2]
	case node.x >= 0 && node.x <= 15 && node.z < 0:
		return neighbours[3]
	case node.x >= 0 && node.x <= 15 && node.z >= 0 && node.z <= 15:
		return centre
	case node.x >= 0 && node.x <= 15 && node.z >= 16:
		return neighbours[4]
	case node.x >= 16 && node.z < 0:
		return neighbours[5]
	case node.x >= 16 && node.z >= 0 && node.z <= 15:
		return neighbours[6]
	case node.x >= 16 && node.z >= 16:
		return neighbours[7]
	}
	panic("should never happen")
}

// highestEmissionLevel checks for the block with the highest emission level at a position and returns it.
func highestEmissionLevel(sub *SubChunk, x, y, z uint8) uint8 {
	storages := sub.storages
	// We offer several fast ways out to get a little more performance out of this.
	switch len(storages) {
	case 0:
		return 0
	case 1:
		id := storages[0].RuntimeID(x, y, z)
		if id == sub.air {
			return 0
		}
		return LightBlocks[id]
	case 2:
		var highest uint8
		id := storages[0].RuntimeID(x, y, z)
		if id != sub.air {
			highest = LightBlocks[id]
		}
		id = storages[1].RuntimeID(x, y, z)
		if id != sub.air {
			if v := LightBlocks[id]; v > highest {
				highest = v
			}
		}
		return highest
	}
	var highest uint8
	for i := range storages {
		if l := LightBlocks[storages[i].RuntimeID(x, y, z)]; l > highest {
			highest = l
		}
	}
	return highest
}

// filterLevel checks for the block with the highest filter level in the sub chunk at a specific position,
// returning 15 if there is a block, but if it is not present in the FilteringBlocks map.
func filterLevel(sub *SubChunk, x, y, z uint8) uint8 {
	storages := sub.storages
	// We offer several fast ways out to get a little more performance out of this.
	switch len(storages) {
	case 0:
		return 0
	case 1:
		id := storages[0].RuntimeID(x, y, z)
		if id == sub.air {
			return 0
		}
		return FilteringBlocks[id]
	case 2:
		var highest uint8

		id := storages[0].RuntimeID(x, y, z)
		if id != sub.air {
			highest = FilteringBlocks[id]
		}

		id = storages[1].RuntimeID(x, y, z)
		if id != sub.air {
			if v := FilteringBlocks[id]; v > highest {
				highest = v
			}
		}
		return highest
	}
	var highest uint8
	for i := range storages {
		id := storages[i].RuntimeID(x, y, z)
		if id != sub.air {
			if l := FilteringBlocks[id]; l > highest {
				highest = l
			}
		}
	}
	return highest
}

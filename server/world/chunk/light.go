package chunk

import (
	"container/list"
	"github.com/df-mc/dragonfly/server/block/cube"
)

// FillLight executes the light 'filling' stage, where the chunk is filled with light coming only from the
// chunk itself, without light crossing chunk borders.
func FillLight(a *Area) {
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
	for cx := 0; cx < a.w; cx++ {
		for cz := 0; cz < a.w; cz++ {
			c := a.c[a.chunkIndex(cx, cz)]
			baseX, baseZ := a.baseX+(cx<<4), a.baseZ+(cz<<4)

			for index, sub := range c.sub {
				// Potential fast path out: We first check the palette to see if there are any blocks that emit light in the
				// block storage. If not, we don't need to iterate the full storage.
				if sub == nil || !anyBlockLight(sub) {
					continue
				}
				baseY := int(subY(int16(index)))
				for y := 0; y < 16; y++ {
					for x := 0; x < 16; x++ {
						for z := 0; z < 16; z++ {
							pos := cube.Pos{x + baseX, y + baseY, z + baseZ}
							if level := a.Highest(pos, LightBlocks); level > 0 {
								queue.PushBack(node(pos, level, BlockLight))
							}
						}
					}
				}
			}
		}
	}
}

// anyBlockLight checks if there are any blocks in the SubChunk passed that emit light.
func anyBlockLight(sub *SubChunk) bool {
	for _, layer := range sub.storages {
		for _, id := range layer.palette.blockRuntimeIDs {
			if LightBlocks[id] != 0 {
				return true
			}
		}
	}
	return false
}

// insertSkyLightNodes iterates over the chunk and inserts a light node anywhere at the highest block in the
// chunk. In addition, any sky light above those nodes will be set to 15.
func insertSkyLightNodes(queue *list.List, a *Area) {
	for cx := 0; cx < a.w; cx++ {
		for cz := 0; cz < a.w; cz++ {
			c := a.c[a.chunkIndex(cx, cz)]
			baseX, baseZ := a.baseX+(cx<<4), a.baseZ+(cz<<4)

			m := calculateHeightmap(c)
			highestY := cube.MinY
			for index := range c.sub {
				if c.sub[index] != nil {
					highestY = int(subY(int16(index)) + 15)
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
						// Fill the rest of the column with sky light on full strength.
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
	for cx := 0; cx < a.w; cx++ {
		for cz := 0; cz < a.w; cz++ {
			c := a.c[a.chunkIndex(cx, cz)]
			baseX, baseZ := a.baseX+(cx<<4), a.baseZ+(cz<<4)

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
							l := lt.light(sub, x, y, z)
							if l <= 1 {
								// The light level was either 0 or 1, meaning it cannot propagate either way.
								continue
							}
							nodeNeeded := false
							if x == 0 {
								subNeighbour := neighbours[1].sub[index]
								if subNeighbour != nil && lt.light(subNeighbour, 15, y, z) < l {
									nodeNeeded = true
								}
							} else if x == 15 {
								subNeighbour := neighbours[6].sub[index]
								if subNeighbour != nil && lt.light(subNeighbour, 0, y, z) < l {
									nodeNeeded = true
								}
							}
							if !nodeNeeded {
								if z == 0 {
									subNeighbour := neighbours[3].sub[index]
									if subNeighbour != nil && lt.light(subNeighbour, x, y, 15) < l {
										nodeNeeded = true
									}
								} else if z == 15 {
									subNeighbour := neighbours[4].sub[index]
									if subNeighbour != nil && lt.light(subNeighbour, x, y, 0) < l {
										nodeNeeded = true
									}
								}
							}
							if nodeNeeded {
								queue.PushBack(lightNode{x: int8(x), y: totalY, z: int8(z), level: l})
							}
						}
					}
				}
			}
		}
	}
}

// propagate propagates the next light node in the node queue passed within the chunk itself. It does not
// spread the light beyond the chunk.
func propagate(queue *list.List, a *Area) {
	n := queue.Remove(queue.Front()).(lightNode)
	a.SetLight(n.pos, n.lt, n.level)

	// If the level is 1 or lower, it won't be able to propagate any further.
	if n.level > 1 {
		for _, neighbour := range a.Neighbours(n) {
			if filter := a.Highest(neighbour.pos, FilteringBlocks) + 1; filter < n.level && a.Light(neighbour.pos, n.lt) < n.level {
				neighbour.level = n.level - filter
				queue.PushBack(neighbour)
			}
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

// filterLevel checks for the block with the Highest filter level in the sub chunk at a specific position,
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

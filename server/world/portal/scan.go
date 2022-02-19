package portal

import (
	"container/list"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// scanIteration contains data about a nether portal scan iteration.
type scanIteration struct {
	lastPos cube.Pos
	face    cube.Face
	first   bool
}

// multiAxisScan performs a scan on the Z and X axis, returning the result that had the most positions, although
// favouring the Z axis.
func multiAxisScan(framePos cube.Pos, w *world.World, matchers []world.Block) (cube.Axis, []cube.Pos, int, int, bool, bool) {
	positions, width, height, completed := scan(cube.Z, framePos, w, matchers)
	positionsTwo, widthTwo, heightTwo, completedTwo := scan(cube.X, framePos, w, matchers)
	if len(positionsTwo) > len(positions) && !completed {
		return cube.X, positionsTwo, widthTwo, heightTwo, completedTwo, len(positionsTwo) > 0
	}
	return cube.Z, positions, width, height, completed, len(positions) > 0
}

// scan performs a scan on the given axis for any of the provided matchers using a position and a world.
func scan(axis cube.Axis, framePos cube.Pos, w *world.World, matchers []world.Block) ([]cube.Pos, int, int, bool) {
	var width, height int
	positionsMap := make(map[cube.Pos]bool)

	completed := true
	queue := list.New()
	queue.PushBack(scanIteration{lastPos: framePos, first: true})
	for queue.Len() > 0 {
		e := queue.Front()
		queue.Remove(e)

		// Parse the latest iteration.
		iteration := e.Value.(scanIteration)
		pos := iteration.lastPos
		if !iteration.first {
			pos = pos.Side(iteration.face)
		}

		b := w.Block(pos)
		if _, ok := positionsMap[pos]; !ok && satisfiesMatchers(b, matchers) {
			// Add the position to the map.
			positionsMap[pos] = true

			// If we are on the same X or Z axis as the portal, we can assume that our height is being changed.
			if pos.X() == framePos.X() && pos.Z() == framePos.Z() && iteration.face < cube.FaceNorth {
				height++
			}

			// If we are on the same Y axis as the portal, we can assume that our width is being changed.
			if pos.Y() == framePos.Y() {
				width++
			}

			// Make sure we don't exceed the maximum portal width or height.
			if width > maximumNetherPortalWidth || height > maximumNetherPortalHeight {
				return []cube.Pos{}, 0, 0, false
			}

			// Plan new iterations.
			if axis == cube.Z {
				queue.PushBack(scanIteration{lastPos: pos, face: cube.FaceSouth})
				queue.PushBack(scanIteration{lastPos: pos, face: cube.FaceNorth})
			} else if axis == cube.X {
				queue.PushBack(scanIteration{lastPos: pos, face: cube.FaceWest})
				queue.PushBack(scanIteration{lastPos: pos, face: cube.FaceEast})
			}
			queue.PushBack(scanIteration{lastPos: pos, face: cube.FaceUp})
			queue.PushBack(scanIteration{lastPos: pos, face: cube.FaceDown})
		} else if _, ok = positionsMap[pos]; !(ok || b == obsidian()) {
			completed = false
		}
	}

	// Make sure we at least reach the minimum portal width and height.
	area, expectedArea := len(positionsMap), width*height
	completed = width >= minimumNetherPortalWidth && height >= minimumNetherPortalHeight && area == expectedArea

	// Get the actual positions from the map.
	positions := make([]cube.Pos, 0, expectedArea)
	for pos := range positionsMap {
		positions = append(positions, pos)
	}
	return positions, width, height, completed
}

// satisfiesMatchers checks if the given block satisfies all matchers.
func satisfiesMatchers(b world.Block, matchers []world.Block) bool {
	for _, matcher := range matchers {
		if b == matcher {
			return true
		}
	}
	return false
}

// air returns an air block.
func air() world.Block {
	a, ok := world.BlockByName("minecraft:air", nil)
	if !ok {
		panic("could not find air block")
	}
	return a
}

// portal returns a portal block.
func portal(axis cube.Axis) world.Block {
	p, ok := world.BlockByName("minecraft:portal", map[string]interface{}{"portal_axis": axis.String()})
	if !ok {
		panic("could not find portal block")
	}
	return p
}

// obsidian returns an obsidian block.
func obsidian() world.Block {
	o, ok := world.BlockByName("minecraft:obsidian", nil)
	if !ok {
		panic("could not find obsidian block")
	}
	return o
}

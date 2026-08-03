package portal

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/world"
)

// blockMatcher reports whether a block belongs to a portal interior on the given axis.
type blockMatcher func(world.Block, cube.Axis) bool

// multiAxisScan performs a scan on the Z and X axis, favouring the Z axis unless only the X axis reaches the minimum
// area. The last return value reports whether a portal-like interior was found; use Framed to check completion.
func multiAxisScan(framePos cube.Pos, tx *world.Tx, matches blockMatcher) (cube.Axis, []cube.Pos, int, int, bool, bool) {
	zPositions, zWidth, zHeight, zCompleted := scan(cube.Z, framePos, tx, matches)
	xPositions, xWidth, xHeight, xCompleted := scan(cube.X, framePos, tx, matches)
	if len(zPositions) < minimumNetherPortalArea && len(xPositions) >= minimumNetherPortalArea {
		return cube.X, xPositions, xWidth, xHeight, xCompleted, len(xPositions) > 0
	}
	return cube.Z, zPositions, zWidth, zHeight, zCompleted, len(zPositions) > 0
}

// scan validates a vertical rectangular portal interior on the given horizontal axis.
func scan(axis cube.Axis, pos cube.Pos, tx *world.Tx, matches blockMatcher) ([]cube.Pos, int, int, bool) {
	// Return if the starting block isn't part of a portal interior.
	if !matches(tx.Block(pos), axis) {
		return nil, 0, 0, false
	}
	negative, positive := axis.Faces()

	// Walk down then towards the negative face to land on the bottom-left interior corner.
	origin := pos
	for down, next := 0, origin.Side(cube.FaceDown); matches(tx.Block(next), axis); down, next = down+1, origin.Side(cube.FaceDown) {
		if down >= maximumNetherPortalHeight {
			return nil, 0, 0, false
		}
		origin = next
	}
	for left, next := 0, origin.Side(negative); matches(tx.Block(next), axis); left, next = left+1, origin.Side(negative) {
		if left >= maximumNetherPortalWidth {
			return nil, 0, 0, false
		}
		origin = next
	}

	// Measure the bottom row and the leftmost column from the origin.
	width := 0
	for p := origin; matches(tx.Block(p), axis); p = p.Side(positive) {
		width++
		if width > maximumNetherPortalWidth {
			return nil, 0, 0, false
		}
	}
	height := 0
	for p := origin; matches(tx.Block(p), axis); p = p.Side(cube.FaceUp) {
		height++
		if height > maximumNetherPortalHeight {
			return nil, 0, 0, false
		}
	}
	// Reject anything smaller than the minimum frame size.
	if width < minimumNetherPortalWidth || height < minimumNetherPortalHeight {
		return nil, width, height, false
	}

	// Validate each row: side frames intact and every interior block matches.
	positions := make([]cube.Pos, 0, width*height)
	for y := 0; y < height; y++ {
		row := origin.Add(cube.Pos{0, y})
		if !isFrame(tx.Block(row.Side(negative))) || !isFrame(tx.Block(row.Add(widthOffset(axis, width)))) {
			return nil, width, height, false
		}
		for x := 0; x < width; x++ {
			p := row.Add(widthOffset(axis, x))
			if !matches(tx.Block(p), axis) {
				return nil, width, height, false
			}
			positions = append(positions, p)
		}
	}
	// Validate the top and bottom frames over each column.
	for x := 0; x < width; x++ {
		p := origin.Add(widthOffset(axis, x))
		if !isFrame(tx.Block(p.Side(cube.FaceDown))) || !isFrame(tx.Block(p.Add(cube.Pos{0, height}))) {
			return nil, width, height, false
		}
	}
	return positions, width, height, true
}

// connectedNetherPortal flood-fills the region of portal blocks reachable from pos and returns its axis and positions.
// Used to clean up an entire portal when its frame breaks, where scan would only return a partial rectangle.
func connectedNetherPortal(tx *world.Tx, pos cube.Pos) (cube.Axis, []cube.Pos, bool) {
	for _, axis := range []cube.Axis{cube.Z, cube.X} {
		if !matchesNetherPortal(tx.Block(pos), axis) {
			continue
		}
		positions := connectedPortalBlocks(tx, pos, axis)
		return axis, positions, len(positions) > 0
	}
	return 0, nil, false
}

// connectedPortalBlocks returns every portal block of the given axis reachable from pos via face neighbours.
func connectedPortalBlocks(tx *world.Tx, pos cube.Pos, axis cube.Axis) []cube.Pos {
	var positions []cube.Pos
	queue := []cube.Pos{pos}
	seen := map[cube.Pos]struct{}{pos: {}}
	faces := portalFaces(axis)
	for len(queue) > 0 {
		p := queue[0]
		queue = queue[1:]
		if !matchesNetherPortal(tx.Block(p), axis) {
			continue
		}
		positions = append(positions, p)
		for _, face := range faces {
			next := p.Side(face)
			if _, ok := seen[next]; ok {
				continue
			}
			seen[next] = struct{}{}
			queue = append(queue, next)
		}
	}
	return positions
}

// portalFaces returns the four neighbouring faces used to flood-fill a portal of the given horizontal axis.
func portalFaces(axis cube.Axis) []cube.Face {
	negative, positive := axis.Faces()
	return []cube.Face{cube.FaceDown, cube.FaceUp, negative, positive}
}

// widthOffset returns the position offset for moving by the given number of blocks along the portal's width axis.
func widthOffset(axis cube.Axis, offset int) cube.Pos {
	if axis == cube.X {
		return cube.Pos{offset, 0, 0}
	}
	return cube.Pos{0, 0, offset}
}

// isFrame reports whether the block can act as a Nether portal frame block.
func isFrame(b world.Block) bool {
	f, ok := b.(frameBlock)
	return ok && f.Frame(world.Nether)
}

// matchesNetherPortalInterior reports whether the block may sit inside an unactivated Nether portal frame.
func matchesNetherPortalInterior(b world.Block, _ cube.Axis) bool {
	i, ok := b.(interface {
		PortalInterior(target world.Dimension) bool
	})
	return ok && i.PortalInterior(world.Nether)
}

// matchesNetherPortal reports whether the block is an active Nether portal block aligned with the given axis.
func matchesNetherPortal(b world.Block, axis cube.Axis) bool {
	p, ok := b.(portalBlock)
	if !ok || p.Portal() != world.Nether {
		return false
	}
	m, ok := b.Model().(model.Portal)
	return ok && m.Axis == axis
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
	p, ok := world.BlockByName("minecraft:portal", map[string]any{"portal_axis": axis.String()})
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

package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Ladder is the model for a ladder block.
type Ladder struct {
	// Facing is the side opposite to the block the Ladder is currently attached to.
	Facing cube.Direction
}

// BBox returns one physics.BBox that depends on the facing direction of the Ladder.
func (l Ladder) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{full.ExtendTowards(l.Facing.Face(), -0.8125)}
}

// FaceSolid always returns false.
func (l Ladder) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

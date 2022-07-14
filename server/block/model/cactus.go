package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Cactus is the model for a Cactus. It is just barely not a full block, having a slightly reduced width and depth.
type Cactus struct{}

// BBox returns a physics.BBox that is slightly smaller than a full block.
func (Cactus) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{cube.Box(0.025, 0, 0.025, 0.975, 1, 0.975)}
}

// FaceSolid always returns false.
func (Cactus) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

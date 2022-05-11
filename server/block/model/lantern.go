package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Lantern is a model for the lantern block. It can be placed on the ground or hanging from the ceiling.
type Lantern struct {
	// Hanging specifies if the lantern is hanging from a block or if it's placed on the ground.
	Hanging bool
}

// BBox returns a physics.BBox attached to either the ceiling or to the ground.
func (l Lantern) BBox(cube.Pos, *world.World) []cube.BBox {
	if l.Hanging {
		return []cube.BBox{cube.Box(0.3125, 0.125, 0.3125, 0.6875, 0.625, 0.6875)}
	}
	return []cube.BBox{cube.Box(0.3125, 0, 0.3125, 0.6875, 0.5, 0.6875)}
}

// FaceSolid always returns false.
func (l Lantern) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

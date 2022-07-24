package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Campfire is the model used by Campfires
type Campfire struct{}

// BBox returns a flat BBox with a width of 0.0625.
func (Campfire) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 0.4375, 1)}
}

// FaceSolid always returns false.
func (Campfire) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

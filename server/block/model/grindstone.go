package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Grindstone is a model used by grindstones.
type Grindstone struct {
	// Axis is the axis the grindstone is attached to.
	Axis cube.Axis
}

// BBox ...
func (g Grindstone) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{cube.Box(0.125, 0.125, 0.125, 0.825, 0.825, 0.825).Stretch(g.Axis, 0.125)}
}

// FaceSolid always returns false.
func (g Grindstone) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

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
func (g Grindstone) BBox(pos cube.Pos, w *world.World) []cube.BBox {
	switch g.Axis {
	case cube.X:
		return []cube.BBox{cube.Box(0, 0.125, 0.125, 1, 0.825, 0.825)}
	case cube.Y:
		return []cube.BBox{cube.Box(0.125, 0, 0.125, 0.825, 1, 0.825)}
	case cube.Z:
		return []cube.BBox{cube.Box(0.125, 0.125, 0, 0.825, 0.825, 1)}
	}
	panic("should never happen")
}

// FaceSolid always returns false.
func (g Grindstone) FaceSolid(pos cube.Pos, face cube.Face, w *world.World) bool {
	return false
}

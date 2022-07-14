package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Carpet is a model for carpet-like extremely thin blocks.
type Carpet struct{}

// BBox returns a flat BBox with a width of 0.0625.
func (Carpet) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 0.0625, 1)}
}

// FaceSolid always returns false.
func (Carpet) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

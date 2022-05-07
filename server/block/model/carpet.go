package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Carpet is a model for carpet-like extremely thin blocks.
type Carpet struct{}

// BBox returns a flat BBox with a width of 0.0625.
func (Carpet) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{cube.Box(mgl64.Vec3{}, mgl64.Vec3{1, 0.0625, 1})}
}

// FaceSolid always returns false.
func (Carpet) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Solid is the model of a fully solid block. Blocks with this model, such as stone or wooden planks, have a
// 1x1x1 collision box.
type Solid struct{}

// BBox returns a physics.BBox spanning a full block.
func (Solid) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{cube.Box(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1})}
}

// FaceSolid always returns true.
func (Solid) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return true
}

package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Solid is the model of a fully solid block. Blocks with this model, such as stone or wooden planks, have a
// 1x1x1 collision box.
type Solid struct{}

// full is a BBox that occupies a full block.
var full = cube.Box(0, 0, 0, 1, 1, 1)

// BBox returns a physics.BBox spanning a full block.
func (Solid) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{full}
}

// FaceSolid always returns true.
func (Solid) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return true
}

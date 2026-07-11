package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Scaffolding is the model of a scaffolding block. It has a solid slab across its top so entities can stand on
// top of it, along with four thin corner posts, leaving the centre open to be climbed through.
type Scaffolding struct{}

// BBox returns the top slab and four corner posts of the Scaffolding.
func (Scaffolding) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{
		cube.Box(0, 0.875, 0, 1, 1, 1),
		cube.Box(0, 0, 0, 0.125, 1, 0.125),
		cube.Box(0.875, 0, 0, 1, 1, 0.125),
		cube.Box(0, 0, 0.875, 0.125, 1, 1),
		cube.Box(0.875, 0, 0.875, 1, 1, 1),
	}
}

// FaceSolid always returns false.
func (Scaffolding) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}

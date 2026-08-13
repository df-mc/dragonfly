package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Honey is the model of a honey block. It is a full block tall, but inset by a pixel on each of its
// horizontal sides, which is what allows entities to slide down against them.
type Honey struct{}

// honey is the 14x14x16 pixel BBox of a honey block.
var honey = cube.Box(1.0/16, 0, 1.0/16, 15.0/16, 1, 15.0/16)

// BBox ...
func (Honey) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{honey}
}

// FaceSolid ...
func (Honey) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}

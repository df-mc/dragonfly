package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Cauldron is a hollow model with a raised floor and four walls.
type Cauldron struct{}

// BBox ...
func (Cauldron) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{
		cube.Box(0, 0, 0, 1, 0.3125, 1),
		cube.Box(0, 0, 0, 0.125, 1, 1),
		cube.Box(0, 0, 0, 1, 1, 0.125),
		cube.Box(0.875, 0, 0, 1, 1, 1),
		cube.Box(0, 0, 0.875, 1, 1, 1),
	}
}

// FaceSolid returns true for every face except the open top.
func (Cauldron) FaceSolid(_ cube.Pos, face cube.Face, _ world.BlockSource) bool {
	return face != cube.FaceUp
}

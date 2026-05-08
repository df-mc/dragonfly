package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Bed is a model used for beds. This model works for both parts of the bed.
type Bed struct{}

// BBox ...
func (b Bed) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 0.5625, 1)}
}

// FaceSolid ...
func (Bed) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}

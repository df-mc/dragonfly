package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Stonecutter is a model used by stonecutters.
type Stonecutter struct{}

// BBox ...
func (Stonecutter) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 0.5625, 1)}
}

// FaceSolid ...
func (Stonecutter) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}

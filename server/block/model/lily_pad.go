package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// LilyPad is a model for the lily pad block.
type LilyPad struct{}

// BBox ...
func (LilyPad) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{cube.Box(0.0625, 0, 0.0625, 0.9375, 0.015625, 0.9375)}
}

// FaceSolid ...
func (LilyPad) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}

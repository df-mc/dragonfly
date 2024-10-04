package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Lectern is a model used by lecterns.
type Lectern struct{}

// BBox ...
func (Lectern) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 0.9, 1)}
}

// FaceSolid ...
func (Lectern) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}

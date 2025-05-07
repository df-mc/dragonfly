package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// SnowLayer is a model used by the snow layer blocks.
type SnowLayer struct {
	// Height is the height of the snow layer. It ranges from 0 to 7.
	Height int
}

// BBox ...
func (s SnowLayer) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	height := 0.5
	if s.Height >= 3 {
		height = 1
	}
	return []cube.BBox{cube.Box(0, 0, 0, 1, height, 1)}
}

// FaceSolid ...
func (s SnowLayer) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return s.Height >= 7
}

package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Chain is a model used by chain blocks.
type Chain struct {
	// Axis is the axis which the chain faces.
	Axis cube.Axis
}

func (c Chain) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{cube.Box(0.40625, 0.40625, 0.40625, 0.59375, 0.59375, 0.59375).Stretch(c.Axis, 0.40625)}
}

func (Chain) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}

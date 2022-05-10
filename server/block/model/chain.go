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

// BBox ...
func (c Chain) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{cube.Box(0.40625, 0.40625, 0.40625, 0.59375, 0.59375, 0.59375).Stretch(c.Axis, 0.40625)}
}

// FaceSolid ...
func (Chain) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

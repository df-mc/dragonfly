package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// HangingSign is a model used by hanging signs. Only a sign mounted on a wall collides, with the bar at its top.
type HangingSign struct {
	// Wall is true if the sign is mounted on the side of a block.
	Wall bool
	// Axis is the axis the bar of a wall mounted sign runs along.
	Axis cube.Axis
}

// BBox ...
func (h HangingSign) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	if !h.Wall {
		return nil
	}
	return []cube.BBox{cube.Box(0.375, 0.875, 0.375, 0.625, 1, 0.625).Stretch(h.Axis, 0.375)}
}

// FaceSolid ...
func (HangingSign) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}

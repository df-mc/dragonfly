package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Anvil is a model used by anvils.
type Anvil struct {
	// Facing is the direction that the anvil is facing.
	Facing cube.Direction
}

// BBox ...
func (a Anvil) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{full.Stretch(a.Facing.RotateLeft().Face().Axis(), -0.125)}
}

// FaceSolid ...
func (Anvil) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

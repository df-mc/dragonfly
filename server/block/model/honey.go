package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Honey is a model used by honey blocks. It is identical to Solid except that its collision box is one
// sixteenth of a block shorter on the top face, matching vanilla's honey block shape.
type Honey struct{}

// BBox returns a physics.BBox spanning a full block except for the topmost 1/16th.
func (Honey) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{full.ExtendTowards(cube.FaceUp, -0.0625)}
}

// FaceSolid always returns true.
func (Honey) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return true
}

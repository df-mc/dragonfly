package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

type Hopper struct{}

// BBox returns a physics.BBox that spans a full block.
func (Hopper) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{full}
}

// FaceSolid only returns true for the top face of the hopper.
func (Hopper) FaceSolid(_ cube.Pos, face cube.Face, _ *world.World) bool {
	return face == cube.FaceUp
}

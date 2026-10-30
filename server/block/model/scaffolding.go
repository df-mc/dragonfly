package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Scaffolding is the model of a scaffolding block. Only entities on top of it collide with it, so it only has a
// slab across the top, and an unsupported scaffolding has no collision at all.
type Scaffolding struct {
	// Unsupported is true if the scaffolding has a stability of 7.
	Unsupported bool
}

// BBox ...
func (s Scaffolding) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	if s.Unsupported {
		return nil
	}
	return []cube.BBox{cube.Box(0, 0.875, 0, 1, 1, 1)}
}

// FaceSolid only returns true for the top face.
func (Scaffolding) FaceSolid(_ cube.Pos, face cube.Face, _ world.BlockSource) bool {
	return face == cube.FaceUp
}

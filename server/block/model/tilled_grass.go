package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// TilledGrass is a model used for grass that has been tilled in some way, such as dirt paths and farmland.
type TilledGrass struct{}

// BBox returns a physics.BBox that spans an entire block.
func (TilledGrass) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{full.ExtendTowards(cube.FaceDown, 0.0625)}
}

// FaceSolid always returns true.
func (TilledGrass) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return true
}

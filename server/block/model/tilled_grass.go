package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/support"
	"github.com/df-mc/dragonfly/server/world"
)

// TilledGrass is a model used for grass that has been tilled in some way, such as dirt paths and farmland.
type TilledGrass struct{}

// SupportType ...
func (TilledGrass) SupportType(cube.Face) support.Type {
	return support.Full{}
}

// BBox returns a physics.BBox that spans an entire block.
func (TilledGrass) BBox(cube.Pos, *world.World) []cube.BBox {
	// TODO: Make the max Y value 0.9375 once https://bugs.mojang.com/browse/MCPE-12109 gets fixed.
	return []cube.BBox{full}
}

// FaceSolid always returns true.
func (TilledGrass) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return true
}

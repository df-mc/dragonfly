package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// TilledGrass is a model used for grass that has been tilled in some way, such as dirt paths and farmland.
type TilledGrass struct{}

// BBox returns a physics.BBox that spans an entire block.
func (TilledGrass) BBox(cube.Pos, *world.World) []cube.BBox {
	// TODO: Make the max Y value 0.9375 once https://bugs.mojang.com/browse/MCPE-12109 gets fixed.
	return []cube.BBox{cube.Box(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1})}
}

// FaceSolid always returns true.
func (TilledGrass) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return true
}

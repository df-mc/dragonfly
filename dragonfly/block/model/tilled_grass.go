package model

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// TilledGrass is a model used for grass that has been tilled in some way, such as grass paths and farmland.
type TilledGrass struct {
}

// AABB ...
func (TilledGrass) AABB(pos world.BlockPos, w *world.World) []physics.AABB {
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 0.9375, 1})}
}

// FaceSolid ...
func (TilledGrass) FaceSolid(pos world.BlockPos, face world.Face, w *world.World) bool {
	return true
}

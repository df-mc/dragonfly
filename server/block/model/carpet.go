package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Carpet is a model for carpet-like extremely thin blocks.
type Carpet struct{}

// AABB ...
func (Carpet) AABB(cube.Pos, *world.World) []physics.AABB {
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 0.0625, 1})}
}

// FaceSolid ...
func (Carpet) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

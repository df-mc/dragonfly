package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// FenceGate is a model used by fence gates.
type FenceGate struct {
	Facing cube.Direction
	Open   bool
}

// AABB ...
func (f FenceGate) AABB(cube.Pos, *world.World) []physics.AABB {
	if f.Open {
		return nil
	}
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1.5, 1}).Stretch(f.Facing.Face().Axis(), -0.375)}
}

// FaceSolid ...
func (f FenceGate) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

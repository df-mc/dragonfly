package model

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Door is a model used for doors. It has no solid faces and a bounding box that changes depending on
// the direction of the door, whether it is open, and the side of its hinge.
type Door struct {
	Facing world.Direction
	Open   bool
	Right  bool
}

// AABB ...
func (d Door) AABB(pos world.BlockPos, w *world.World) []physics.AABB {
	if d.Open {
		if d.Right {
			return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}).ExtendTowards(int(d.Facing.Rotate90().Opposite().Face()), -0.8125)}
		}
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}).ExtendTowards(int(d.Facing.Rotate90().Face()), -0.8125)}
	}
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}).ExtendTowards(int(d.Facing.Face()), -0.8125)}
}

// FaceSolid ...
func (d Door) FaceSolid(pos world.BlockPos, face world.Face, w *world.World) bool {
	return false
}

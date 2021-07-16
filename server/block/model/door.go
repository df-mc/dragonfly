package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Door is a model used for doors. It has no solid faces and a bounding box that changes depending on
// the direction of the door, whether it is open, and the side of its hinge.
type Door struct {
	Facing cube.Direction
	Open   bool
	Right  bool
}

// AABB ...
func (d Door) AABB(pos cube.Pos, w *world.World) []physics.AABB {
	if d.Open {
		if d.Right {
			return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}).ExtendTowards(d.Facing.RotateLeft().Face(), -0.8125)}
		}
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}).ExtendTowards(d.Facing.RotateRight().Face(), -0.8125)}
	}
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}).ExtendTowards(d.Facing.Face(), -0.8125)}
}

// FaceSolid ...
func (d Door) FaceSolid(pos cube.Pos, face cube.Face, w *world.World) bool {
	return false
}

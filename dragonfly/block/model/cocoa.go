package model

import (
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Cocoa is a model used by cocoa bean blocks.
type Cocoa struct {
	Facing world.Direction
	Age    int
}

// AABB ...
func (c Cocoa) AABB(pos world.BlockPos, w *world.World) []physics.AABB {
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1}).
		Stretch(int(c.Facing.Rotate90().Face().Axis()), -((6-float64(c.Age))/16)).
		ExtendTowards(int(world.FaceDown), -((7-float64(c.Age)*2)/16)).
		ExtendTowards(int(world.FaceUp), -0.25).
		ExtendTowards(int(c.Facing.Face()), -0.0625).
		ExtendTowards(int(c.Facing.Opposite().Face()), -((11 - float64(c.Age)*2) / 16))}
}

// FaceSolid ...
func (c Cocoa) FaceSolid(pos world.BlockPos, face world.Face, w *world.World) bool {
	return false
}

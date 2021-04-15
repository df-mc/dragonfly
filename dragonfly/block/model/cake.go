package model

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Cake is a model used by cake blocks.
type Cake struct {
	Bites int
}

// AABB ...
func (c Cake) AABB(pos cube.Pos, w *world.World) []physics.AABB {
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{0.0625, 0, 0.0625}, mgl64.Vec3{0.9375, 0.5, 0.9375}).
		ExtendTowards(cube.FaceWest, -(float64(c.Bites) / 8))}
}

// FaceSolid ...
func (c Cake) FaceSolid(pos cube.Pos, face cube.Face, w *world.World) bool {
	return false
}

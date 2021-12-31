package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Cake is a model used by cake blocks.
type Cake struct {
	// Bites is the amount of bites that were taken from the cake. A cake can have up to 7 bites taken from it, before
	// being consumed entirely.
	Bites int
}

// AABB returns an AABB with a size that depends on the amount of bites taken.
func (c Cake) AABB(cube.Pos, *world.World) []physics.AABB {
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{0.0625, 0, 0.0625}, mgl64.Vec3{0.9375, 0.5, 0.9375}).
		ExtendTowards(cube.FaceWest, -(float64(c.Bites) / 8))}
}

// FaceSolid always returns false.
func (c Cake) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

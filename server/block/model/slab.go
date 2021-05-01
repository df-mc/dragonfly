package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Slab is the model of a slab-like block, which is either a half block or a full block, depending on if the
// slab is double.
type Slab struct {
	Double, Top bool
}

// AABB ...
func (s Slab) AABB(cube.Pos, *world.World) []physics.AABB {
	if s.Double {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1})}
	}
	if s.Top {
		return []physics.AABB{physics.NewAABB(mgl64.Vec3{0, 0.5, 0}, mgl64.Vec3{1, 1, 1})}
	}
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 0.5, 1})}
}

// FaceSolid ...
func (s Slab) FaceSolid(_ cube.Pos, face cube.Face, _ *world.World) bool {
	if s.Double {
		return true
	} else if s.Top {
		return face == cube.FaceUp
	}
	return face == cube.FaceDown
}

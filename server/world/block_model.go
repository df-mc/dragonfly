package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/entity/physics"
	"github.com/go-gl/mathgl/mgl64"
)

// BlockModel represents the model of a block. These models specify the ways a block can be collided with and
// whether or not specific faces are solid wrt. being able to, for example, place torches onto those sides.
type BlockModel interface {
	// AABB returns the bounding boxes that a block with this model can be collided with.
	AABB(pos cube.Pos, w *World) []physics.AABB
	// FaceSolid checks if a specific face of a block at the position in a world passed is solid. Blocks may
	// be attached to these faces.
	FaceSolid(pos cube.Pos, face cube.Face, w *World) bool
}

// unknownModel is the model used for unknown blocks. It is the equivalent of a fully solid model.
type unknownModel struct{}

// AABB ...
func (u unknownModel) AABB(cube.Pos, *World) []physics.AABB {
	return []physics.AABB{physics.NewAABB(mgl64.Vec3{}, mgl64.Vec3{1, 1, 1})}
}

// FaceSolid ...
func (u unknownModel) FaceSolid(cube.Pos, cube.Face, *World) bool {
	return true
}

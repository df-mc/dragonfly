package world

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/entity/physics"
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

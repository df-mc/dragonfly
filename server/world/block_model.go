package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
)

// BlockModel represents the model of a block. These models specify the ways a block can be collided with and
// whether specific faces are solid wrt. being able to, for example, place torches onto those sides.
type BlockModel interface {
	// BBox returns the bounding boxes that a block with this model can be collided with.
	BBox(pos cube.Pos, w *World) []cube.BBox
	// FaceSolid checks if a specific face of a block at the position in a world passed is solid. Blocks may
	// be attached to these faces.
	FaceSolid(pos cube.Pos, face cube.Face, w *World) bool
}

// unknownModel is the model used for unknown blocks. It is the equivalent of a fully solid model.
type unknownModel struct{}

// BBox ...
func (u unknownModel) BBox(cube.Pos, *World) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 1, 1)}
}

// FaceSolid ...
func (u unknownModel) FaceSolid(cube.Pos, cube.Face, *World) bool {
	return true
}

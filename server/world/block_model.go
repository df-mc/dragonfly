package world

import (
	"github.com/df-mc/dragonfly/server/block/cube"
)

// BlockModel represents the model of a block. These models specify the ways a block can be collided with and
// whether specific faces are solid wrt. being able to, for example, place torches onto those sides.
type BlockModel interface {
	// BBox returns the bounding boxes that a block with this model can be collided with.
	BBox(pos cube.Pos, s BlockSource) []cube.BBox
	// FaceSolid checks if a specific face of a block at the position in a world passed is solid. Blocks may
	// be attached to these faces.
	FaceSolid(pos cube.Pos, face cube.Face, s BlockSource) bool
}

// unknownModel is the model used for unknown blocks. It is the equivalent of a fully solid model.
type unknownModel struct{}

func (u unknownModel) BBox(cube.Pos, BlockSource) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 1, 1)}
}

func (u unknownModel) FaceSolid(cube.Pos, cube.Face, BlockSource) bool {
	return true
}

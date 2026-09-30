package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// ShelfMushroom is a model used by shelf mushroom blocks. It is a thin plate on the side of the block it is
// attached to, reaching further out once the mushroom has grown.
type ShelfMushroom struct {
	// Facing is the direction the shelf mushroom faces, away from the block it is attached to.
	Facing cube.Direction
	// Large is true if the shelf mushroom has grown to its large stage.
	Large bool
}

// BBox returns a plate hugging the side behind the mushroom.
func (s ShelfMushroom) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	depth := 0.5
	if s.Large {
		depth = 0.75
	}
	return []cube.BBox{full.
		ExtendTowards(cube.FaceUp, -0.75).
		ExtendTowards(s.Facing.Face(), -(1 - depth))}
}

// FaceSolid always returns false.
func (ShelfMushroom) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}

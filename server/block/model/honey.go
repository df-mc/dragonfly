package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Honey is the model of a honey block. Unlike a fully solid block, its collision box is inset by one
// pixel (1/16) on every horizontal side and is one pixel shorter than a full block, giving it a 14x14x15
// pixel hitbox. This inset allows entities pressed up against its sides to slide down it.
type Honey struct{}

// honeyBox is the 14x14x15 pixel collision box of a honey block.
var honeyBox = cube.Box(1.0/16, 0, 1.0/16, 15.0/16, 15.0/16, 15.0/16)

// BBox returns the 14x14x15 pixel bounding box of a honey block.
func (Honey) BBox(cube.Pos, world.BlockSource) []cube.BBox {
	return []cube.BBox{honeyBox}
}

// FaceSolid always returns false, since a honey block is not a full block on any of its faces.
func (Honey) FaceSolid(cube.Pos, cube.Face, world.BlockSource) bool {
	return false
}

package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/support"
	"github.com/df-mc/dragonfly/server/world"
)

// Leaves is a model for leaves-like blocks. These blocks have a full collision box, but none of their faces
// are solid.
type Leaves struct{}

// SupportType ...
func (Leaves) SupportType(cube.Face) support.Type {
	return support.None{}
}

// BBox returns a physics.BBox that spans a full block.
func (Leaves) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{full}
}

// FaceSolid always returns false.
func (Leaves) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

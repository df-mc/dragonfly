package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/support"
	"github.com/df-mc/dragonfly/server/world"
)

// Empty is a model that is completely empty. It has no collision boxes or solid faces.
type Empty struct{}

// SupportType ...
func (Empty) SupportType(cube.Face) support.Type {
	return support.None{}
}

// BBox returns an empty slice.
func (Empty) BBox(cube.Pos, *world.World) []cube.BBox {
	return nil
}

// FaceSolid always returns false.
func (Empty) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

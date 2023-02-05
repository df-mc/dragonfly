package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Diode is a model used by redstone gates, such as repeaters and comparators.
type Diode struct{}

// BBox ...
func (Diode) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{full.ExtendTowards(cube.FaceDown, 0.875)}
}

// FaceSolid ...
func (Diode) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

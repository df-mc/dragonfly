package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// Lectern is a model used by lecterns.
type Lectern struct{}

// BBox ...
func (Lectern) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{full.ExtendTowards(cube.FaceDown, 0.1)}
}

// FaceSolid ...
func (Lectern) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

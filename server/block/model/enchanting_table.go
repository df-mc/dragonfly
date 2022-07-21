package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// EnchantingTable is a model used by enchanting tables.
type EnchantingTable struct{}

// BBox ...
func (EnchantingTable) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{full.ExtendTowards(cube.FaceDown, 0.25)}
}

// FaceSolid ...
func (EnchantingTable) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

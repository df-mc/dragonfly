package model

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
)

// EnchantingTable is a model used by enchanting tables.
type EnchantingTable struct{}

// BBox ...
func (EnchantingTable) BBox(cube.Pos, *world.World) []cube.BBox {
	return []cube.BBox{cube.Box(0, 0, 0, 1, 0.75, 1)}
}

// FaceSolid ...
func (EnchantingTable) FaceSolid(cube.Pos, cube.Face, *world.World) bool {
	return false
}

package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// SoulSoil is a block naturally found only in the soul sand valley.
type SoulSoil struct {
	solid
}

// BreakInfo ...
func (s SoulSoil) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.5,
		Harvestable: alwaysHarvestable,
		Effective:   shovelEffective,
		Drops:       simpleDrops(item.NewStack(s, 1)),
	}
}

// EncodeItem ...
func (SoulSoil) EncodeItem() (name string, meta int16) {
	return "minecraft:soul_soil", 0
}

// EncodeBlock ...
func (SoulSoil) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:soul_soil", nil
}

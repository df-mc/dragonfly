package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// CoalOre is a common ore.
type CoalOre struct {
	solid
	bassDrum
}

// BreakInfo ...
func (c CoalOre) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(item.Coal{}, 1)), //TODO: Silk Touch
		XPDrops:     XPDropRange{0, 2},
	}
}

// EncodeItem ...
func (CoalOre) EncodeItem() (name string, meta int16) {
	return "minecraft:coal_ore", 0
}

// EncodeBlock ...
func (CoalOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:coal_ore", nil
}

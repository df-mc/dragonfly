package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// EmeraldOre is an ore generating exclusively under mountain biomes.
type EmeraldOre struct {
	noNBT
	solid
	bassDrum
}

// BreakInfo ...
func (e EmeraldOre) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 3,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierIron.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(item.Emerald{}, 1)), //TODO: Silk Touch
		XPDrops:   XPDropRange{3, 7},
	}
}

// EncodeItem ...
func (EmeraldOre) EncodeItem() (id int32, meta int16) {
	return 129, 0
}

// EncodeBlock ...
func (EmeraldOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:emerald_ore", nil
}

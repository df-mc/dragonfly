package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
)

// GoldOre is a rare mineral block found underground.
type GoldOre struct {
	solid
	bassDrum
}

// BreakInfo ...
func (g GoldOre) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 3,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierIron.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(g, 1)),
	}
}

// EncodeItem ...
func (GoldOre) EncodeItem() (name string, meta int16) {
	return "minecraft:gold_ore", 0
}

// EncodeBlock ...
func (GoldOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:gold_ore", nil
}

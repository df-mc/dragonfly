package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
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
func (GoldOre) EncodeItem() (id int32, name string, meta int16) {
	return 14, "minecraft:gold_ore", 0
}

// EncodeBlock ...
func (GoldOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:gold_ore", nil
}

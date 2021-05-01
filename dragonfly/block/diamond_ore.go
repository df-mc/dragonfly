package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// DiamondOre is a rare ore that generates underground.
type DiamondOre struct {
	solid
	bassDrum
}

// BreakInfo ...
func (d DiamondOre) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 3,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierIron.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(item.Diamond{}, 1)), //TODO: Silk Touch
		XPDrops:   XPDropRange{3, 7},
	}
}

// EncodeItem ...
func (DiamondOre) EncodeItem() (id int32, name string, meta int16) {
	return 56, "minecraft:diamond_ore", 0
}

// EncodeBlock ...
func (DiamondOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:diamond_ore", nil
}

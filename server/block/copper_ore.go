package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
)

// CopperOre is a rare mineral block found underground.
type CopperOre struct {
	solid
	bassDrum
}

// BreakInfo ...
func (c CopperOre) BreakInfo() BreakInfo {
	return newBreakInfo(3, func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierStone.HarvestLevel
	}, pickaxeEffective, oneOf(item.RawCopper{}))
}

// EncodeItem ...
func (c CopperOre) EncodeItem() (name string, meta int16) {
	return "minecraft:copper_ore", 0
}

// EncodeBlock ...
func (c CopperOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:copper_ore", nil
}

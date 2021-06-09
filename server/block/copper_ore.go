package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
)

// CopperOre is a rare mineral block found underground.
type CopperOre struct {
	solid
	bassDrum

	// Type is the type of copper ore.
	Type OreType
}

// BreakInfo ...
func (c CopperOre) BreakInfo() BreakInfo {
	return newBreakInfo(c.Type.Hardness(), func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierStone.HarvestLevel
	}, pickaxeEffective, oneOf(item.RawCopper{})) //TODO: Silk Touch
}

// EncodeItem ...
func (c CopperOre) EncodeItem() (name string, meta int16) {
	switch c.Type {
	case StoneOre():
		return "minecraft:copper_ore", 0
	case DeepslateOre():
		return "minecraft:deepslate_copper_ore", 0
	}
	panic("unknown ore type")
}

// EncodeBlock ...
func (c CopperOre) EncodeBlock() (string, map[string]interface{}) {
	switch c.Type {
	case StoneOre():
		return "minecraft:copper_ore", nil
	case DeepslateOre():
		return "minecraft:deepslate_copper_ore", nil
	}
	panic("unknown ore type")
}

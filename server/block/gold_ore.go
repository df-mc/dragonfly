package block

import (
	"github.com/df-mc/dragonfly/server/item/tool"
)

// GoldOre is a rare mineral block found underground.
type GoldOre struct {
	solid
	bassDrum

	// Type is the type of gold ore.
	Type OreType
}

// BreakInfo ...
func (g GoldOre) BreakInfo() BreakInfo {
	return newBreakInfo(g.Type.Hardness(), func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierIron.HarvestLevel
	}, pickaxeEffective, oneOf(g))
}

// EncodeItem ...
func (g GoldOre) EncodeItem() (name string, meta int16) {
	switch g.Type {
	case StoneOre():
		return "minecraft:gold_ore", 0
	case DeepslateOre():
		return "minecraft:deepslate_gold_ore", 0
	}
	panic("unknown ore type")
}

// EncodeBlock ...
func (g GoldOre) EncodeBlock() (string, map[string]interface{}) {
	switch g.Type {
	case StoneOre():
		return "minecraft:gold_ore", nil
	case DeepslateOre():
		return "minecraft:deepslate_gold_ore", nil
	}
	panic("unknown ore type")
}

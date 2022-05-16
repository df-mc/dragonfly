package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// EmeraldOre is an ore generating exclusively under mountain biomes.
type EmeraldOre struct {
	solid
	bassDrum

	// Type is the type of emerald ore.
	Type OreType
}

// BreakInfo ...
func (e EmeraldOre) BreakInfo() BreakInfo {
	return newBreakInfo(e.Type.Hardness(), func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierIron.HarvestLevel
	}, pickaxeEffective, silkTouchOneOf(item.Emerald{}, e)).withXPDropRange(3, 7)
}

// EncodeItem ...
func (e EmeraldOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + e.Type.Prefix() + "emerald_ore", 0
}

// EncodeBlock ...
func (e EmeraldOre) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + e.Type.Prefix() + "emerald_ore", nil
}

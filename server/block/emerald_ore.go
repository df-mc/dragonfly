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
	i := newBreakInfo(e.Type.Hardness(), func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.TierIron.HarvestLevel
	}, pickaxeEffective, silkTouchOneOf(item.Emerald{}, e))
	i.XPDrops = XPDropRange{3, 7}
	return i
}

// EncodeItem ...
func (e EmeraldOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + e.Type.Prefix() + "emerald_ore", 0
}

// EncodeBlock ...
func (e EmeraldOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:" + e.Type.Prefix() + "emerald_ore", nil
}

package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
)

// EmeraldOre is an ore generating exclusively under mountain biomes.
type EmeraldOre struct {
	solid
	bassDrum
}

// BreakInfo ...
func (e EmeraldOre) BreakInfo() BreakInfo {
	// TODO: Silk touch.
	i := newBreakInfo(3, func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierIron.HarvestLevel
	}, pickaxeEffective, oneOf(item.Emerald{}))
	i.XPDrops = XPDropRange{3, 7}
	return i
}

// EncodeItem ...
func (EmeraldOre) EncodeItem() (name string, meta int16) {
	return "minecraft:emerald_ore", 0
}

// EncodeBlock ...
func (EmeraldOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:emerald_ore", nil
}

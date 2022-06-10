package block

import (
	"github.com/df-mc/dragonfly/server/item"
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
	return newBreakInfo(g.Type.Hardness(), func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierIron.HarvestLevel
	}, pickaxeEffective, silkTouchOneOf(item.RawGold{}, g))
}

// EncodeItem ...
func (g GoldOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + g.Type.Prefix() + "gold_ore", 0
}

// EncodeBlock ...
func (g GoldOre) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + g.Type.Prefix() + "gold_ore", nil
}

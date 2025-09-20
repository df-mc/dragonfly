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

func (e EmeraldOre) BreakInfo() BreakInfo {
	i := newBreakInfo(e.Type.Hardness(), func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierIron.HarvestLevel
	}, pickaxeEffective, silkTouchOneOf(item.Emerald{}, e)).withXPDropRange(3, 7)
	if e.Type == DeepslateOre() {
		i = i.withBlastResistance(15)
	}
	return i
}

func (EmeraldOre) SmeltInfo() item.SmeltInfo {
	return newOreSmeltInfo(item.NewStack(item.Emerald{}, 1), 1)
}

func (e EmeraldOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + e.Type.Prefix() + "emerald_ore", 0
}

func (e EmeraldOre) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + e.Type.Prefix() + "emerald_ore", nil
}

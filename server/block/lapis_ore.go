package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"math/rand/v2"
)

// LapisOre is an ore block from which lapis lazuli is obtained.
type LapisOre struct {
	solid
	bassDrum

	// Type is the type of lapis ore.
	Type OreType
}

// BreakInfo ...
func (l LapisOre) BreakInfo() BreakInfo {
	i := newBreakInfo(l.Type.Hardness(), func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierStone.HarvestLevel
	}, pickaxeEffective, silkTouchDrop(item.NewStack(item.LapisLazuli{}, rand.IntN(5)+4), item.NewStack(l, 1))).withXPDropRange(2, 5)
	if l.Type == DeepslateOre() {
		i = i.withBlastResistance(9)
	}
	return i
}

// SmeltInfo ...
func (LapisOre) SmeltInfo() item.SmeltInfo {
	return newOreSmeltInfo(item.NewStack(item.LapisLazuli{}, 1), 0.2)
}

// EncodeItem ...
func (l LapisOre) EncodeItem() (name string, meta int16) {
	return "minecraft:" + l.Type.Prefix() + "lapis_ore", 0
}

// EncodeBlock ...
func (l LapisOre) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + l.Type.Prefix() + "lapis_ore", nil
}

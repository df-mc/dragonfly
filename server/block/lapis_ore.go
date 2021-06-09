package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
	"math/rand"
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
	// TODO: Silk touch.
	i := newBreakInfo(l.Type.Hardness(), func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierStone.HarvestLevel
	}, pickaxeEffective, simpleDrops(item.NewStack(item.LapisLazuli{}, rand.Intn(5)+4)))
	i.XPDrops = XPDropRange{2, 5}
	return i
}

// EncodeItem ...
func (l LapisOre) EncodeItem() (name string, meta int16) {
	switch l.Type {
	case StoneOre():
		return "minecraft:lapis_ore", 0
	case DeepslateOre():
		return "minecraft:deepslate_lapis_ore", 0
	}
	panic("unknown ore type")
}

// EncodeBlock ...
func (l LapisOre) EncodeBlock() (string, map[string]interface{}) {
	switch l.Type {
	case StoneOre():
		return "minecraft:lapis_ore", nil
	case DeepslateOre():
		return "minecraft:deepslate_lapis_ore", nil
	}
	panic("unknown ore type")
}

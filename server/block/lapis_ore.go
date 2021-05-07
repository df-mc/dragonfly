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
}

// BreakInfo ...
func (l LapisOre) BreakInfo() BreakInfo {
	// TODO: Silk touch.
	i := newBreakInfo(3, func(t tool.Tool) bool {
		return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierStone.HarvestLevel
	}, pickaxeEffective, simpleDrops(item.NewStack(item.LapisLazuli{}, rand.Intn(5)+4)))
	i.XPDrops = XPDropRange{2, 5}
	return i
}

// EncodeItem ...
func (LapisOre) EncodeItem() (name string, meta int16) {
	return "minecraft:lapis_ore", 0
}

// EncodeBlock ...
func (LapisOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:lapis_ore", nil
}

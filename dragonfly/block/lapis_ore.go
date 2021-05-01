package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"math/rand"
)

// LapisOre is an ore block from which lapis lazuli is obtained.
type LapisOre struct {
	solid
	bassDrum
}

// BreakInfo ...
func (l LapisOre) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 3,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierStone.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(item.LapisLazuli{}, rand.Intn(5)+4)), //TODO: Silk Touch
		XPDrops:   XPDropRange{2, 5},
	}
}

// EncodeItem ...
func (LapisOre) EncodeItem() (id int32, name string, meta int16) {
	return 21, "minecraft:lapis_ore", 0
}

// EncodeBlock ...
func (LapisOre) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:lapis_ore", nil
}

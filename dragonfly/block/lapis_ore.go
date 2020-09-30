package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"math/rand"
)

// LapisOre is an ore block from which lapis lazuli is obtained.
type LapisOre struct {
	noNBT
	solid
	bassdrum
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
func (l LapisOre) EncodeItem() (id int32, meta int16) {
	return 21, 0
}

// EncodeBlock ...
func (l LapisOre) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:lapis_ore", nil
}

// Hash ...
func (l LapisOre) Hash() uint64 {
	return hashLapisOre
}

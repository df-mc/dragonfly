package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// IronOre is a mineral block found underground.
type IronOre struct {
	noNBT
	solid
}

// BreakInfo ...
func (i IronOre) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 3,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierStone.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(i, 1)),
	}
}

// EncodeItem ...
func (i IronOre) EncodeItem() (id int32, meta int16) {
	return 15, 0
}

// EncodeBlock ...
func (i IronOre) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:iron_ore", nil
}

// Hash ...
func (i IronOre) Hash() uint64 {
	return hashIronOre
}

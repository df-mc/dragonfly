package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// Prismarine Block is a Decorative Block, commonly found in Ocean Monnuments
type Prismarine struct {
	noNBT
	solid
}

// BreakInfo ...
func (p Prismarine) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 1.5,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierWood.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(p, 1)),
	}
}

// EncodeItem ...
func (p Prismarine) EncodeItem() (id int32, meta int16) {
	return 168, 0
}

// EncodeBlock ...
func (p Prismarine) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:prismarine", nil
}

// Hash ...
func (p Prismarine) Hash() uint64 {
	return hashPrismarine
}

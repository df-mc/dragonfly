package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// AncientDebris is a rare ore found within The Nether.
type AncientDebris struct {
	noNBT
	solid
}

// BreakInfo ...
func (a AncientDebris) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 30,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierDiamond.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(a, 1)),
	}
}

// EncodeItem ...
func (a AncientDebris) EncodeItem() (id int32, meta int16) {
	return -271, 0
}

// EncodeBlock ...
func (a AncientDebris) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:ancient_debris", nil
}

// Hash ...
func (a AncientDebris) Hash() uint64 {
	return hashAncientDebris
}

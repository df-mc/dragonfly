package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
)

// AncientDebris is a rare ore found within The Nether.
type AncientDebris struct {
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
func (AncientDebris) EncodeItem() (id int32, name string, meta int16) {
	return -271, "minecraft:ancient_debris", 0
}

// EncodeBlock ...
func (AncientDebris) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:ancient_debris", nil
}

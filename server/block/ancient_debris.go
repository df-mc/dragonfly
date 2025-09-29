package block

import (
	"github.com/df-mc/dragonfly/server/item"
)

// AncientDebris is a rare ore found within The Nether.
type AncientDebris struct {
	solid
}

func (a AncientDebris) BreakInfo() BreakInfo {
	return newBreakInfo(30, func(t item.Tool) bool {
		return t.ToolType() == item.TypePickaxe && t.HarvestLevel() >= item.ToolTierDiamond.HarvestLevel
	}, pickaxeEffective, oneOf(a)).withBlastResistance(6000)
}

func (AncientDebris) SmeltInfo() item.SmeltInfo {
	return newOreSmeltInfo(item.NewStack(item.NetheriteScrap{}, 1), 2)
}

func (AncientDebris) EncodeItem() (name string, meta int16) {
	return "minecraft:ancient_debris", 0
}

func (AncientDebris) EncodeBlock() (string, map[string]any) {
	return "minecraft:ancient_debris", nil
}

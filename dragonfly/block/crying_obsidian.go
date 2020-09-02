package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// CryingObsidian is a luminous variant of obsidian that can be used to craft a respawn anchor and produces purple particles when placed.
type CryingObsidian struct {
	noNBT
	solid
}

// LightEmissionLevel ...
func (CryingObsidian) LightEmissionLevel() uint8 {
	return 10
}

// BreakInfo ...
func (c CryingObsidian) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 50,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() == tool.TierDiamond.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(c, 1)),
	}
}

// EncodeItem ...
func (CryingObsidian) EncodeItem() (id int32, meta int16) {
	return -289, 0
}

// EncodeBlock ...
func (CryingObsidian) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:crying_obsidian", nil
}

// Hash ...
func (CryingObsidian) Hash() uint64 {
	return hashCryingObsidian
}

package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

// Obsidian is a dark purple block known for its high blast resistance and strength, most commonly found when
// water flows over lava.
type Obsidian struct {
	solid
	bassDrum
}

// EncodeItem ...
func (Obsidian) EncodeItem() (id int32, name string, meta int16) {
	return 49, "minecraft:obsidian", 0
}

// EncodeBlock ...
func (Obsidian) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:obsidian", nil
}

// BreakInfo ...
func (o Obsidian) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness: 50,
		Harvestable: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypePickaxe && t.HarvestLevel() >= tool.TierDiamond.HarvestLevel
		},
		Effective: pickaxeEffective,
		Drops:     simpleDrops(item.NewStack(o, 1)),
	}
}

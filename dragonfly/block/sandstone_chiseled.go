package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// SandstoneChiseled is a block which naturally generates in Desert Villages and Desert Temples.
type SandstoneChiseled struct {
	// Red specifies if the sandstone is red or not. SandstoneChiseled only has it's basic colour and red.
	Red bool
}

// BreakInfo ...
func (s SandstoneChiseled) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.8,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(s, 1)),
	}
}

// EncodeItem ...
func (s SandstoneChiseled) EncodeItem() (id int32, meta int16) {
	if s.Red {
		return 179, 1
	}
	return 24, 1
}

// EncodeBlock ...
func (s SandstoneChiseled) EncodeBlock() (name string, properties map[string]interface{}) {
	var blockName = "minecraft:sandstone"
	if s.Red {
		blockName = "minecraft:red_sandstone"
	}
	return blockName, map[string]interface{}{"sand_stone_type": "heiroglyphs"}
}

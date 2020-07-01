package block

import "github.com/df-mc/dragonfly/dragonfly/item"

// CutSandstone is a block which can be found in generated structures.
type CutSandstone struct {
	// Red specifies if the sandstone is red or not. ChiseledSandstone only has it's basic colour and red.
	Red bool
}

// BreakInfo ...
func (s CutSandstone) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.8,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(s, 1)),
	}
}

// EncodeItem ...
func (s CutSandstone) EncodeItem() (id int32, meta int16) {
	if s.Red {
		return 179, 2
	}
	return 24, 2
}

// EncodeBlock ...
func (s CutSandstone) EncodeBlock() (name string, properties map[string]interface{}) {
	var blockName = "minecraft:sandstone"
	if s.Red {
		blockName = "minecraft:red_sandstone"
	}
	return blockName, map[string]interface{}{"sand_stone_type": "cut"}
}

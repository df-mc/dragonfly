package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

//Sand is a block which can be found in a desert or on beaches.
type Sandstone struct {
	// Red specifies if the sandstone is red or not. Sandstone only has it's basic colour and red.
	Red bool
	// Smooth specifies the block state of Sandstone
	Smooth bool
}

// BreakInfo ...
func (s Sandstone) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.8,
		Harvestable: pickaxeHarvestable,
		Effective:   pickaxeEffective,
		Drops:       simpleDrops(item.NewStack(s, 1)),
	}
}

// EncodeItem ...
func (s Sandstone) EncodeItem() (id int32, meta int16) {
	var m int16 = 0
	if s.Smooth {
		m = 1
	}
	if s.Red {
		return 179, m
	}
	return 24, m
}

// EncodeBlock ...
func (s Sandstone) EncodeBlock() (name string, properties map[string]interface{}) {
	var blockName = "minecraft:sandstone"
	if s.Red {
		blockName = "minecraft:red_sandstone"
	}
	if s.Smooth {
		return blockName, map[string]interface{}{"sand_stone_type": "smooth"}
	}
	return blockName, map[string]interface{}{"sand_stone_type": "default"}
}

// allSandstone returns all states of sandstone.
func allSandstone() (sandstone []world.Block) {
	f := func(red bool, smooth bool) {
		sandstone = append(sandstone, Sandstone{Smooth: smooth, Red: red})
	}
	f(false, false)
	f(false, true)
	f(true, false)
	f(true, true)
	return
}

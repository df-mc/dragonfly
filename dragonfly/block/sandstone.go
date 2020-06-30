package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

//Sand is a block which can be found in a desert or on beaches.
type Sandstone struct {
	// ColourRed specifies if the sandstone is red or not. Sandstone only has it's basic colour and red.
	ColourRed bool
	// DataValue specifies the block state of the sandstone.
	//Valid values: 0 (Default), 1 (Chiseled), 2 (Cut), 3 (Smooth)
	DataValue int16
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
	if s.ColourRed {
		return 179, s.DataValue
	}
	return 24, s.DataValue
}

// EncodeBlock ...
func (s Sandstone) EncodeBlock() (name string, properties map[string]interface{}) {
	var blockName = "minecraft:sandstone"
	if s.ColourRed {
		blockName = "minecraft:red_sandstone"
	}
	switch s.DataValue {
	case 0:
		return blockName, map[string]interface{}{"sand_stone_type": "default"}
	case 1:
		return blockName, map[string]interface{}{"sand_stone_type": "heiroglyphs"}
	case 2:
		return blockName, map[string]interface{}{"sand_stone_type": "cut"}
	case 3:
		return blockName, map[string]interface{}{"sand_stone_type": "smooth"}
	}
	panic("invalid sandstone type")
}

// allSandstone returns a list of all possible sandstone states.
func allSandstone() (sandstone []world.Block) {
	f := func(colour bool) {
		sandstone = append(sandstone, Sandstone{ColourRed: colour, DataValue: 0})
		sandstone = append(sandstone, Sandstone{ColourRed: colour, DataValue: 1})
		sandstone = append(sandstone, Sandstone{ColourRed: colour, DataValue: 2})
		sandstone = append(sandstone, Sandstone{ColourRed: colour, DataValue: 3})
	}
	f(true)
	f(false)
	return
}

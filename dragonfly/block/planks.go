package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block/material"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
)

// Planks are common blocks used in crafting recipes. They are made by crafting logs into planks.
type Planks struct {
	// Wood is the type of wood of the planks. This field must have one of the values found in the material
	// package.
	Wood material.Wood
}

// BreakInfo ...
func (p Planks) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(item.NewStack(p, 1)),
	}
}

// EncodeItem ...
func (p Planks) EncodeItem() (id int32, meta int16) {
	switch p.Wood {
	case material.OakWood():
		return 5, 0
	case material.SpruceWood():
		return 5, 1
	case material.BirchWood():
		return 5, 2
	case material.JungleWood():
		return 5, 3
	case material.AcaciaWood():
		return 5, 4
	case material.DarkOakWood():
		return 5, 5
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (p Planks) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:planks", map[string]interface{}{"wood_type": p.Wood.String()}
}

// allPlanks returns all planks types.
func allPlanks() []world.Block {
	return []world.Block{
		Planks{Wood: material.OakWood()},
		Planks{Wood: material.SpruceWood()},
		Planks{Wood: material.BirchWood()},
		Planks{Wood: material.JungleWood()},
		Planks{Wood: material.AcaciaWood()},
		Planks{Wood: material.DarkOakWood()},
	}
}

package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// Planks are common blocks used in crafting recipes. They are made by crafting logs into planks.
type Planks struct {
	noNBT
	// Wood is the type of wood of the planks. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
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
	case wood.Oak():
		return 5, 0
	case wood.Spruce():
		return 5, 1
	case wood.Birch():
		return 5, 2
	case wood.Jungle():
		return 5, 3
	case wood.Acacia():
		return 5, 4
	case wood.DarkOak():
		return 5, 5
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (p Planks) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:planks", map[string]interface{}{"wood_type": p.Wood.String()}
}

// Hash ...
func (p Planks) Hash() uint64 {
	return hashPlanks | (uint64(p.Wood.Uint8()) << 32)
}

// allPlanks returns all planks types.
func allPlanks() []world.Block {
	return []world.Block{
		Planks{Wood: wood.Oak()},
		Planks{Wood: wood.Spruce()},
		Planks{Wood: wood.Birch()},
		Planks{Wood: wood.Jungle()},
		Planks{Wood: wood.Acacia()},
		Planks{Wood: wood.DarkOak()},
	}
}

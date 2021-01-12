package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// Planks are common blocks used in crafting recipes. They are made by crafting logs into planks.
type Planks struct {
	noNBT
	solid
	bass

	// Wood is the type of wood of the planks. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
}

// FlammabilityInfo ...
func (p Planks) FlammabilityInfo() FlammabilityInfo {
	if !p.Wood.Flammable() {
		return FlammabilityInfo{}
	}
	return FlammabilityInfo{
		Encouragement: 5,
		Flammability:  20,
		LavaFlammable: true,
	}
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
	case wood.Crimson():
		return -242, 0
	case wood.Warped():
		return -243, 0
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (p Planks) EncodeBlock() (name string, properties map[string]interface{}) {
	switch p.Wood {
	case wood.Crimson():
		return "minecraft:crimson_planks", nil
	case wood.Warped():
		return "minecraft:warped_planks", nil
	default:
		return "minecraft:planks", map[string]interface{}{"wood_type": p.Wood.String()}
	}
}

// Hash ...
func (p Planks) Hash() uint64 {
	return hashPlanks | (uint64(p.Wood.Uint8()) << 32)
}

// allPlanks returns all planks types.
func allPlanks() (planks []world.Block) {
	for _, w := range wood.All() {
		planks = append(planks, Planks{Wood: w})
	}
	return
}

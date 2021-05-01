package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// Planks are common blocks used in crafting recipes. They are made by crafting logs into planks.
type Planks struct {
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
func (p Planks) EncodeItem() (id int32, name string, meta int16) {
	switch p.Wood {
	case wood.Oak(), wood.Spruce(), wood.Birch(), wood.Jungle(), wood.Acacia(), wood.DarkOak():
		return 5, "minecraft:planks", int16(p.Wood.Uint8())
	case wood.Crimson():
		return -242, "minecraft:" + p.Wood.String() + "_planks", 0
	case wood.Warped():
		return -243, "minecraft:" + p.Wood.String() + "_planks", 0
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

// allPlanks returns all planks types.
func allPlanks() (planks []world.Block) {
	for _, w := range wood.All() {
		planks = append(planks, Planks{Wood: w})
	}
	return
}

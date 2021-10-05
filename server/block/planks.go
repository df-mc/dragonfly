package block

import (
	"github.com/df-mc/dragonfly/server/world"
)

// Planks are common blocks used in crafting recipes. They are made by crafting logs into planks.
type Planks struct {
	solid
	bass

	// Wood is the type of wood of the planks. This field must have one of the values found in the material
	// package.
	Wood WoodType
}

// FlammabilityInfo ...
func (p Planks) FlammabilityInfo() FlammabilityInfo {
	if !p.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(5, 20, true)
}

// BreakInfo ...
func (p Planks) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, oneOf(p))
}

// EncodeItem ...
func (p Planks) EncodeItem() (name string, meta int16) {
	switch p.Wood {
	case OakWood(), SpruceWood(), BirchWood(), JungleWood(), AcaciaWood(), DarkOakWood():
		return "minecraft:planks", int16(p.Wood.Uint8())
	case CrimsonWood(), WarpedWood():
		return "minecraft:" + p.Wood.String() + "_planks", 0
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (p Planks) EncodeBlock() (name string, properties map[string]interface{}) {
	switch p.Wood {
	case CrimsonWood(), WarpedWood():
		return "minecraft:" + p.Wood.String() + "_planks", nil
	default:
		return "minecraft:planks", map[string]interface{}{"wood_type": p.Wood.String()}
	}
}

// allPlanks returns all planks types.
func allPlanks() (planks []world.Block) {
	for _, w := range WoodTypes() {
		planks = append(planks, Planks{Wood: w})
	}
	return
}

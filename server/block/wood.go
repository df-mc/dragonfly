package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Wood is a block that has the log's "bark" texture on all six sides. It comes in 8 types: oak, spruce, birch, jungle,
// acacia, dark oak, crimson, and warped.
// Stripped wood is a variant obtained by using an axe on the wood.
type Wood struct {
	solid
	bass

	// Wood is the type of wood.
	Wood WoodType
	// Stripped specifies if the wood is stripped or not.
	Stripped bool
	// Axis is the axis which the wood block faces.
	Axis cube.Axis
}

// FlammabilityInfo ...
func (w Wood) FlammabilityInfo() FlammabilityInfo {
	if !w.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(5, 5, true)
}

// UseOnBlock ...
func (w Wood) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, wo *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(wo, pos, face, w)
	if !used {
		return
	}
	w.Axis = face.Axis()

	place(wo, pos, w, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (w Wood) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, oneOf(w))
}

// Strip ...
func (w Wood) Strip() (world.Block, bool) {
	return Wood{Axis: w.Axis, Wood: w.Wood, Stripped: true}, !w.Stripped
}

// EncodeItem ...
func (w Wood) EncodeItem() (name string, meta int16) {
	switch w.Wood {
	case OakWood(), SpruceWood(), BirchWood(), JungleWood(), AcaciaWood(), DarkOakWood():
		if w.Stripped {
			return "minecraft:wood", int16(8 + w.Wood.Uint8())
		}
		return "minecraft:wood", int16(w.Wood.Uint8())
	case CrimsonWood(), WarpedWood():
		if w.Stripped {
			return "minecraft:stripped_" + w.Wood.String() + "_hyphae", 0
		}
		return "minecraft:" + w.Wood.String() + "_hyphae", 0
	default:
		if w.Stripped {
			return "minecraft:stripped_" + w.Wood.String() + "_wood", 0
		}
		return "minecraft:" + w.Wood.String() + "_wood", 0
	}
}

// EncodeBlock ...
func (w Wood) EncodeBlock() (name string, properties map[string]any) {
	switch w.Wood {
	case OakWood(), SpruceWood(), BirchWood(), JungleWood(), AcaciaWood(), DarkOakWood():
		return "minecraft:wood", map[string]any{"wood_type": w.Wood.String(), "pillar_axis": w.Axis.String(), "stripped_bit": boolByte(w.Stripped)}
	case CrimsonWood(), WarpedWood():
		if w.Stripped {
			return "minecraft:stripped_" + w.Wood.String() + "_hyphae", map[string]any{"pillar_axis": w.Axis.String()}
		}
		return "minecraft:" + w.Wood.String() + "_hyphae", map[string]any{"pillar_axis": w.Axis.String()}
	default:
		if w.Stripped {
			return "minecraft:stripped_" + w.Wood.String() + "_wood", map[string]any{"pillar_axis": w.Axis.String()}
		}
		return "minecraft:" + w.Wood.String() + "_wood", map[string]any{"pillar_axis": w.Axis.String(), "stripped_bit": uint8(0)}
	}
}

// allWood returns a list of all possible wood states.
func allWood() (wood []world.Block) {
	for _, w := range WoodTypes() {
		for axis := cube.Axis(0); axis < 3; axis++ {
			wood = append(wood, Wood{Axis: axis, Stripped: true, Wood: w})
			wood = append(wood, Wood{Axis: axis, Stripped: false, Wood: w})
		}
	}
	return
}

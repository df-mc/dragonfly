package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// Log is a naturally occurring block found in trees, primarily used to create planks. It comes in six
// species: oak, spruce, birch, jungle, acacia, and dark oak.
// Stripped log is a variant obtained by using an axe on a log.
type Log struct {
	solid
	bass

	// Wood is the type of wood of the log. This field must have one of the values found in the material
	// package.
	Wood WoodType
	// Stripped specifies if the log is stripped or not.
	Stripped bool
	// Axis is the axis which the log block faces.
	Axis cube.Axis
}

// FlammabilityInfo ...
func (l Log) FlammabilityInfo() FlammabilityInfo {
	if !l.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(5, 5, true)
}

// BreakInfo ...
func (l Log) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, oneOf(l))
}

// SmeltInfo ...
func (Log) SmeltInfo() item.SmeltInfo {
	return newSmeltInfo(item.NewStack(item.Charcoal{}, 1), 0.15)
}

// FuelInfo ...
func (l Log) FuelInfo() item.FuelInfo {
	if !l.Wood.Flammable() {
		return item.FuelInfo{}
	}
	return newFuelInfo(time.Second * 15)
}

// UseOnBlock handles the rotational placing of logs.
func (l Log) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(tx, pos, face, l)
	if !used {
		return
	}
	l.Axis = face.Axis()

	place(tx, pos, l, user, ctx)
	return placed(ctx)
}

// Strip ...
func (l Log) Strip() (world.Block, world.Sound, bool) {
	return Log{Axis: l.Axis, Wood: l.Wood, Stripped: true}, nil, !l.Stripped
}

// EncodeItem ...
func (l Log) EncodeItem() (name string, meta int16) {
	if !l.Stripped {
		switch l.Wood {
		case CrimsonWood(), WarpedWood():
			return "minecraft:" + l.Wood.String() + "_stem", 0
		default:
			return "minecraft:" + l.Wood.String() + "_log", 0
		}
	}
	switch l.Wood {
	case CrimsonWood(), WarpedWood():
		return "minecraft:stripped_" + l.Wood.String() + "_stem", 0
	default:
		return "minecraft:stripped_" + l.Wood.String() + "_log", 0
	}
}

// EncodeBlock ...
func (l Log) EncodeBlock() (name string, properties map[string]any) {
	if !l.Stripped {
		switch l.Wood {
		case CrimsonWood(), WarpedWood():
			return "minecraft:" + l.Wood.String() + "_stem", map[string]any{"pillar_axis": l.Axis.String()}
		default:
			return "minecraft:" + l.Wood.String() + "_log", map[string]any{"pillar_axis": l.Axis.String()}
		}
	}
	switch l.Wood {
	case CrimsonWood(), WarpedWood():
		return "minecraft:stripped_" + l.Wood.String() + "_stem", map[string]any{"pillar_axis": l.Axis.String()}
	default:
		return "minecraft:stripped_" + l.Wood.String() + "_log", map[string]any{"pillar_axis": l.Axis.String()}
	}
}

// allLogs returns a list of all possible log states.
func allLogs() (logs []world.Block) {
	for _, w := range WoodTypes() {
		for axis := cube.Axis(0); axis < 3; axis++ {
			logs = append(logs, Log{Axis: axis, Stripped: true, Wood: w})
			logs = append(logs, Log{Axis: axis, Stripped: false, Wood: w})
		}
	}
	return
}

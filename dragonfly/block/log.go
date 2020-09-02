package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Log is a naturally occurring block found in trees, primarily used to create planks. It comes in six
// species: oak, spruce, birch, jungle, acacia, and dark oak.
// Stripped log is a variant obtained by using an axe on a log.
type Log struct {
	noNBT
	solid

	// Wood is the type of wood of the log. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
	// Stripped specifies if the log is stripped or not.
	Stripped bool
	// Axis is the axis which the log block faces.
	Axis world.Axis
}

// FlammabilityInfo ...
func (l Log) FlammabilityInfo() FlammabilityInfo {
	return FlammabilityInfo{
		Encouragement: 5,
		Flammability:  5,
		LavaFlammable: true,
	}
}

// UseOnBlock handles the rotational placing of logs.
func (l Log) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, face, used = firstReplaceable(w, pos, face, l)
	if !used {
		return
	}
	l.Axis = face.Axis()

	place(w, pos, l, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (l Log) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    2,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(item.NewStack(l, 1)),
	}
}

// EncodeItem ...
func (l Log) EncodeItem() (id int32, meta int16) {
	switch l.Wood {
	case wood.Oak():
		if l.Stripped {
			return -10, 0
		}
		return 17, 0
	case wood.Spruce():
		if l.Stripped {
			return -5, 0
		}
		return 17, 1
	case wood.Birch():
		if l.Stripped {
			return -6, 0
		}
		return 17, 2
	case wood.Jungle():
		if l.Stripped {
			return -7, 0
		}
		return 17, 3
	case wood.Acacia():
		if l.Stripped {
			return -8, 0
		}
		return 162, 0
	case wood.DarkOak():
		if l.Stripped {
			return -9, 0
		}
		return 162, 1
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (l Log) EncodeBlock() (name string, properties map[string]interface{}) {
	if !l.Stripped {
		switch l.Wood {
		case wood.Oak(), wood.Spruce(), wood.Birch(), wood.Jungle():
			return "minecraft:log", map[string]interface{}{"pillar_axis": l.Axis.String(), "old_log_type": l.Wood.String()}
		case wood.Acacia(), wood.DarkOak():
			return "minecraft:log2", map[string]interface{}{"pillar_axis": l.Axis.String(), "new_log_type": l.Wood.String()}
		}
	}
	switch l.Wood {
	case wood.Oak():
		return "minecraft:stripped_oak_log", map[string]interface{}{"pillar_axis": l.Axis.String()}
	case wood.Spruce():
		return "minecraft:stripped_spruce_log", map[string]interface{}{"pillar_axis": l.Axis.String()}
	case wood.Birch():
		return "minecraft:stripped_birch_log", map[string]interface{}{"pillar_axis": l.Axis.String()}
	case wood.Jungle():
		return "minecraft:stripped_jungle_log", map[string]interface{}{"pillar_axis": l.Axis.String()}
	case wood.Acacia():
		return "minecraft:stripped_acacia_log", map[string]interface{}{"pillar_axis": l.Axis.String()}
	case wood.DarkOak():
		return "minecraft:stripped_dark_oak_log", map[string]interface{}{"pillar_axis": l.Axis.String()}
	}
	panic("invalid wood type")
}

// Hash ...
func (l Log) Hash() uint64 {
	return hashLog | (uint64(boolByte(l.Stripped)) << 32) | (uint64(l.Axis) << 33) | (uint64(l.Wood.Uint8()) << 35)
}

// allLogs returns a list of all possible log states.
func allLogs() (logs []world.Block) {
	f := func(axis world.Axis, stripped bool) {
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: wood.Oak()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: wood.Spruce()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: wood.Birch()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: wood.Jungle()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: wood.Acacia()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: wood.DarkOak()})
	}
	for axis := world.Axis(0); axis < 3; axis++ {
		f(axis, true)
		f(axis, false)
	}
	return
}

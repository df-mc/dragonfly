package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block/material"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
	"github.com/go-gl/mathgl/mgl32"
)

// Log is a naturally occurring block found in trees, primarily used to create planks. It comes in six
// species: oak, spruce, birch, jungle, acacia, and dark oak.
// Stripped log is a variant obtained by using an axe on a log.
type Log struct {
	// Wood is the type of wood of the log. This field must have one of the values found in the material
	// package.
	Wood material.Wood
	// Stripped specifies if the log is stripped or not.
	Stripped bool
	// Axis is the axis which the log block faces.
	Axis world.Axis
}

// UseOnBlock ...
func (l Log) UseOnBlock(pos world.BlockPos, face world.Face, _ mgl32.Vec3, w *world.World, _ item.User, ctx *item.UseContext) bool {
	if replaceable(w, pos.Side(face), l) {
		l.Axis = face.Axis()
		w.PlaceBlock(pos.Side(face), l)

		ctx.SubtractFromCount(1)
		return true
	}
	return false
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
	case material.OakWood():
		if l.Stripped {
			return -10, 0
		}
		return 17, 0
	case material.SpruceWood():
		if l.Stripped {
			return -5, 0
		}
		return 17, 1
	case material.BirchWood():
		if l.Stripped {
			return -6, 0
		}
		return 17, 2
	case material.JungleWood():
		if l.Stripped {
			return -7, 0
		}
		return 17, 3
	case material.AcaciaWood():
		if l.Stripped {
			return -8, 0
		}
		return 162, 0
	case material.DarkOakWood():
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
		case material.OakWood(), material.SpruceWood(), material.BirchWood(), material.JungleWood():
			return "minecraft:log", map[string]interface{}{"pillar_axis": l.Axis.String(), "old_log_type": l.Wood.String()}
		case material.AcaciaWood(), material.DarkOakWood():
			return "minecraft:log2", map[string]interface{}{"pillar_axis": l.Axis.String(), "new_log_type": l.Wood.String()}
		}
	}
	switch l.Wood {
	case material.OakWood():
		return "minecraft:stripped_oak_log", map[string]interface{}{"pillar_axis": l.Axis.String()}
	case material.SpruceWood():
		return "minecraft:stripped_spruce_log", map[string]interface{}{"pillar_axis": l.Axis.String()}
	case material.BirchWood():
		return "minecraft:stripped_birch_log", map[string]interface{}{"pillar_axis": l.Axis.String()}
	case material.JungleWood():
		return "minecraft:stripped_jungle_log", map[string]interface{}{"pillar_axis": l.Axis.String()}
	case material.AcaciaWood():
		return "minecraft:stripped_acacia_log", map[string]interface{}{"pillar_axis": l.Axis.String()}
	case material.DarkOakWood():
		return "minecraft:stripped_dark_oak_log", map[string]interface{}{"pillar_axis": l.Axis.String()}
	}
	panic("invalid wood type")
}

// allLogs returns a list of all possible log states.
func allLogs() (logs []world.Block) {
	f := func(axis world.Axis, stripped bool) {
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: material.OakWood()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: material.SpruceWood()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: material.BirchWood()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: material.JungleWood()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: material.AcaciaWood()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: material.DarkOakWood()})
	}
	for axis := world.Axis(0); axis < 3; axis++ {
		f(axis, true)
		f(axis, false)
	}
	return
}

package block

import (
	"github.com/dragonfly-tech/dragonfly/dragonfly/block/material"
)

// Log is a naturally occurring block found in trees, primarily used to create planks. It comes in six
// species: oak, spruce, birch, jungle, acacia, and dark oak.
// Stripped log is a variant obtained by using an axe on a log.
type Log struct {
	// Wood is the type of wood of the log. This field must have one of the values found in the material
	// package. Using Log without a Wood type will panic.
	Wood material.Wood
	// Stripped specifies if the log is stripped or not.
	Stripped bool

	// Axis is the axis which the log block faces.
	Axis Axis
}

func (l Log) EncodeItem() (id int32, meta int16) {
	if !l.Stripped {
		switch l.Wood {
		case material.OakWood():
			return 17, 0
		case material.SpruceWood():
			return 17, 1
		case material.BirchWood():
			return 17, 2
		case material.JungleWood():
			return 17, 3
		case material.AcaciaWood():
			return 162, 0
		case material.DarkOakWood():
			return 162, 1
		}
	}
	switch l.Wood {
	case material.OakWood():
		return 255 - 265, 0
	case material.SpruceWood():
		return 255 - 260, 0
	case material.BirchWood():
		return 255 - 261, 0
	case material.JungleWood():
		return 255 - 262, 0
	case material.AcaciaWood():
		return 255 - 263, 0
	case material.DarkOakWood():
		return 255 - 264, 0
	}
	panic("invalid wood type")
}

// Name returns the name of the log, including the wood type and whether it is stripped or not.
func (l Log) Name() (name string) {
	if l.Wood == nil {
		panic("log has no wood type")
	}
	if l.Stripped {
		return "Stripped " + l.Wood.Name() + " Log"
	}
	return l.Wood.Name() + " Log"
}

func (l Log) Minecraft() (name string, properties map[string]interface{}) {
	if !l.Stripped {
		switch l.Wood {
		case material.OakWood(), material.SpruceWood(), material.BirchWood(), material.JungleWood():
			return "minecraft:log", map[string]interface{}{"pillar_axis": l.Axis.String(), "old_log_type": l.Wood.Minecraft()}
		case material.AcaciaWood(), material.DarkOakWood():
			return "minecraft:log2", map[string]interface{}{"pillar_axis": l.Axis.String(), "new_log_type": l.Wood.Minecraft()}
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
func allLogs() (logs []Block) {
	f := func(axis Axis, stripped bool) {
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: material.OakWood()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: material.SpruceWood()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: material.BirchWood()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: material.JungleWood()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: material.AcaciaWood()})
		logs = append(logs, Log{Axis: axis, Stripped: stripped, Wood: material.DarkOakWood()})
	}
	for axis := Axis(0); axis < 3; axis++ {
		f(axis, true)
		f(axis, false)
	}
	return
}

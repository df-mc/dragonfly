package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
)

//Sand is a block which can be found in a desert or on beaches.
type Sand struct {
	// ColourRed specifies if the sand is red or not. Sand only has it's basic colour and red.
	Red bool
}

// BreakInfo ...
func (s Sand) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.5,
		Harvestable: alwaysHarvestable,
		Effective: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypeShovel
		},
		Drops: simpleDrops(item.NewStack(s, 1)),
	}
}

// EncodeItem ...
func (s Sand) EncodeItem() (id int32, meta int16) {
	if s.Red {
		return 12, 1
	}
	return 12, 0
}

// EncodeBlock ...
func (s Sand) EncodeBlock() (name string, properties map[string]interface{}) {
	if s.Red {
		return "minecraft:red_sand", map[string]interface{}{"sand_type": "red"}
	}
	return "minecraft:sand", map[string]interface{}{"sand_type": "normal"}
}

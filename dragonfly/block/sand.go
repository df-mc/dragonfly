package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/cube"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/world"
)

// Sand is a block affected by gravity. It can come in a red variant.
type Sand struct {
	gravityAffected
	solid
	snare

	// Red toggles the red sand variant.
	Red bool
}

// NeighbourUpdateTick ...
func (s Sand) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	s.fall(s, pos, w)
}

// BreakInfo ...
func (s Sand) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.5,
		Harvestable: alwaysHarvestable,
		Effective:   shovelEffective,
		Drops:       simpleDrops(item.NewStack(s, 1)),
	}
}

// EncodeItem ...
func (s Sand) EncodeItem() (id int32, name string, meta int16) {
	if s.Red {
		return 12, "minecraft:sand", 1
	}
	return 12, "minecraft:sand", 0
}

// EncodeBlock ...
func (s Sand) EncodeBlock() (string, map[string]interface{}) {
	if s.Red {
		return "minecraft:sand", map[string]interface{}{"sand_type": "red"}
	}
	return "minecraft:sand", map[string]interface{}{"sand_type": "normal"}
}

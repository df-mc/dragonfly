package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Sand is a block affected by gravity. It can come in a red variant.
type Sand struct {
	gravityAffected
	solid
	snare

	// Red toggles the red sand variant.
	Red bool
}

func (s Sand) SoilFor(block world.Block) bool {
	switch block.(type) {
	case Cactus, DeadBush, SugarCane:
		return true
	}
	return false
}

func (s Sand) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	s.fall(s, pos, tx)
}

func (s Sand) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, shovelEffective, oneOf(s))
}

func (Sand) SmeltInfo() item.SmeltInfo {
	return newSmeltInfo(item.NewStack(Glass{}, 1), 0.1)
}

func (s Sand) EncodeItem() (name string, meta int16) {
	if s.Red {
		return "minecraft:red_sand", 0
	}
	return "minecraft:sand", 0
}

func (s Sand) EncodeBlock() (string, map[string]any) {
	if s.Red {
		return "minecraft:red_sand", nil
	}
	return "minecraft:sand", nil
}

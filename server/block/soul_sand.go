package block

import (
	"github.com/df-mc/dragonfly/server/block/instrument"
	"github.com/df-mc/dragonfly/server/item"
)

// SoulSand is a block found naturally only in the Nether. SoulSand slows movement of mobs & players.
type SoulSand struct {
	solid
}

// Instrument ...
func (s SoulSand) Instrument() instrument.Instrument {
	return instrument.CowBell()
}

//TODO: Bubble Columns

// BreakInfo ...
func (s SoulSand) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.5,
		Harvestable: alwaysHarvestable,
		Effective:   shovelEffective,
		Drops:       simpleDrops(item.NewStack(s, 1)),
	}
}

// EncodeItem ...
func (SoulSand) EncodeItem() (id int32, name string, meta int16) {
	return 88, "minecraft:soul_sand", 0
}

// EncodeBlock ...
func (SoulSand) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:soul_sand", nil
}

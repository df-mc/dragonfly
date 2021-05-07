package block

import (
	"github.com/df-mc/dragonfly/server/block/instrument"
)

// SoulSand is a block found naturally only in the Nether. SoulSand slows movement of mobs & players.
type SoulSand struct {
	solid
}

// TODO: Implement bubble columns.

// Instrument ...
func (s SoulSand) Instrument() instrument.Instrument {
	return instrument.CowBell()
}

// BreakInfo ...
func (s SoulSand) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, shovelEffective, oneOf(s))
}

// EncodeItem ...
func (SoulSand) EncodeItem() (name string, meta int16) {
	return "minecraft:soul_sand", 0
}

// EncodeBlock ...
func (SoulSand) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:soul_sand", nil
}

package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/instrument"
	"github.com/df-mc/dragonfly/dragonfly/item"
)

// SoulSand is a block found naturally only in the Nether. SoulSand slows movement of mobs & players.
type SoulSand struct {
	noNBT
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
func (s SoulSand) EncodeItem() (id int32, meta int16) {
	return 88, 0
}

// EncodeBlock ...
func (s SoulSand) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:soul_sand", nil
}

// Hash ...
func (s SoulSand) Hash() uint64 {
	return hashSoulSand
}

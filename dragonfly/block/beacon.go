package block

import (
	"github.com/df-mc/dragonfly/dragonfly/item"
)

// Beacon is a block that projects a light beam skyward, and can provide status effects such as Speed, Jump
// Boost, Haste, Regeneration, Resistance, or Strength to nearby players.
type Beacon struct{ nbt }

// TODO: Implement beacons properly.

// BreakInfo ...
func (b Beacon) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    3,
		Harvestable: alwaysHarvestable,
		Effective:   nothingEffective,
		Drops:       simpleDrops(item.NewStack(b, 1)),
	}
}

// EncodeItem ...
func (b Beacon) EncodeItem() (id int32, meta int16) {
	return 138, 0
}

// EncodeBlock ...
func (b Beacon) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:beacon", nil
}

// Hash ...
func (Beacon) Hash() uint64 {
	return hashBeacon
}

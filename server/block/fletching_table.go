package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"time"
)

// FletchingTable is a block in villages that turn an unemployed villager into a Fletcher.
type FletchingTable struct {
	solid
	bass
}

// BreakInfo ...
func (f FletchingTable) BreakInfo() BreakInfo {
	return newBreakInfo(2.5, alwaysHarvestable, axeEffective, silkTouchOnlyDrop(f))
}

// FuelInfo ...
func (FletchingTable) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// EncodeItem ...
func (FletchingTable) EncodeItem() (string, int16) {
	return "minecraft:fletching_table", 0
}

// EncodeBlock ...
func (FletchingTable) EncodeBlock() (string, map[string]any) {
	return "minecraft:fletching_table", nil
}

package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"time"
)

// FletchingTable is a block in villages that turn an unemployed villager into a Fletcher.
type FletchingTable struct {
	solid
	bassDrum
}

// BreakInfo ...
func (f FletchingTable) BreakInfo() BreakInfo {
	return newBreakInfo(2.5, alwaysHarvestable, axeEffective, silkTouchOnlyDrop(f))
}

// FlammabilityInfo ...
func (FletchingTable) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(5, 5, true)
}

// FuelInfo ...
func (FletchingTable) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// EncodeItem ...
func (FletchingTable) EncodeItem() (name string, meta int16) {
	return "minecraft:fletching_table", 0
}

// EncodeBlock ...
func (FletchingTable) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:fletching_table", nil
}

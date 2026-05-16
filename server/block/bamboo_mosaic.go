package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"time"
)

// BambooMosaic is a decorative bamboo plank variant.
type BambooMosaic struct {
	solid
	bass
}

// FlammabilityInfo ...
func (BambooMosaic) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(5, 20, true)
}

// BreakInfo ...
func (b BambooMosaic) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, oneOf(b)).withBlastResistance(15)
}

// RepairsWoodTools ...
func (BambooMosaic) RepairsWoodTools() bool {
	return true
}

// FuelInfo ...
func (BambooMosaic) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// EncodeItem ...
func (BambooMosaic) EncodeItem() (name string, meta int16) {
	return "minecraft:bamboo_mosaic", 0
}

// EncodeBlock ...
func (BambooMosaic) EncodeBlock() (string, map[string]any) {
	return "minecraft:bamboo_mosaic", nil
}
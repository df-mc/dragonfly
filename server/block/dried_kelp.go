package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"time"
)

// DriedKelp is a block primarily used as fuel in furnaces.
type DriedKelp struct {
	solid
}

// BreakInfo ...
func (d DriedKelp) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, hoeEffective, oneOf(d))
}

// FlammabilityInfo ...
func (DriedKelp) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(30, 60, false)
}

// FuelInfo ...
func (DriedKelp) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 200)
}

// EncodeItem ...
func (DriedKelp) EncodeItem() (name string, meta int16) {
	return "minecraft:dried_kelp_block", 0
}

// EncodeBlock ...
func (DriedKelp) EncodeBlock() (string, map[string]any) {
	return "minecraft:dried_kelp_block", nil
}

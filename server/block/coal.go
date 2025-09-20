package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"time"
)

// Coal is a precious mineral block made from 9 coal.
type Coal struct {
	solid
	bassDrum
}

func (c Coal) BreakInfo() BreakInfo {
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(c)).withBlastResistance(30)
}

func (Coal) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(5, 5, false)
}

func (Coal) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 800)
}

func (Coal) EncodeItem() (name string, meta int16) {
	return "minecraft:coal_block", 0
}

func (Coal) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:coal_block", nil
}

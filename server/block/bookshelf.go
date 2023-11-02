package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"time"
)

// Bookshelf is a decorative block that primarily serves to enhance enchanting with an enchanting table.
type Bookshelf struct {
	solid
	bass
}

// BreakInfo ...
func (b Bookshelf) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, alwaysHarvestable, axeEffective, silkTouchDrop(item.NewStack(item.Book{}, 3), item.NewStack(b, 1)))
}

// FlammabilityInfo ...
func (Bookshelf) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(30, 20, true)
}

// FuelInfo ...
func (Bookshelf) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// EncodeItem ...
func (Bookshelf) EncodeItem() (name string, meta int16) {
	return "minecraft:bookshelf", 0
}

// EncodeBlock ...
func (Bookshelf) EncodeBlock() (string, map[string]any) {
	return "minecraft:bookshelf", nil
}

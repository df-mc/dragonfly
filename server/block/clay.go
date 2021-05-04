package block

import (
	"github.com/df-mc/dragonfly/server/block/instrument"
	"github.com/df-mc/dragonfly/server/item"
)

// Clay is a block that can be found underwater.
type Clay struct {
	solid
}

// Instrument ...
func (c Clay) Instrument() instrument.Instrument {
	return instrument.Flute()
}

// BreakInfo ...
func (c Clay) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.6,
		Harvestable: alwaysHarvestable,
		Effective:   shovelEffective,
		Drops:       simpleDrops(item.NewStack(item.ClayBall{}, 4)), //TODO: Drops itself if mined with silk touch
	}
}

// EncodeItem ...
func (c Clay) EncodeItem() (name string, meta int16) {
	return "minecraft:clay", 0
}

// EncodeBlock ...
func (c Clay) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:clay", nil
}

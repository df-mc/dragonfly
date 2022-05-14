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
	return newBreakInfo(0.6, alwaysHarvestable, shovelEffective, silkTouchDrop(item.NewStack(item.ClayBall{}, 4), item.NewStack(c, 1)), XPDropRange{})
}

// EncodeItem ...
func (c Clay) EncodeItem() (name string, meta int16) {
	return "minecraft:clay", 0
}

// EncodeBlock ...
func (c Clay) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:clay", nil
}

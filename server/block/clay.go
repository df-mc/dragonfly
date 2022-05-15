package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Clay is a block that can be found underwater.
type Clay struct {
	solid
}

// Instrument ...
func (c Clay) Instrument() sound.Instrument {
	return sound.Flute()
}

// BreakInfo ...
func (c Clay) BreakInfo() BreakInfo {
	return newBreakInfo(0.6, alwaysHarvestable, shovelEffective, silkTouchDrop(item.NewStack(item.ClayBall{}, 4), item.NewStack(c, 1)))
}

// EncodeItem ...
func (c Clay) EncodeItem() (name string, meta int16) {
	return "minecraft:clay", 0
}

// EncodeBlock ...
func (c Clay) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:clay", nil
}

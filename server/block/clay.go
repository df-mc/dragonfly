package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Clay is a block that can be found underwater.
type Clay struct {
	solid
}

func (c Clay) Instrument() sound.Instrument {
	return sound.Flute()
}

func (c Clay) BreakInfo() BreakInfo {
	return newBreakInfo(0.6, alwaysHarvestable, shovelEffective, silkTouchDrop(item.NewStack(item.ClayBall{}, 4), item.NewStack(c, 1)))
}

func (Clay) SmeltInfo() item.SmeltInfo {
	return newSmeltInfo(item.NewStack(Terracotta{}, 1), 0.35)
}

func (c Clay) EncodeItem() (name string, meta int16) {
	return "minecraft:clay", 0
}

func (c Clay) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:clay", nil
}

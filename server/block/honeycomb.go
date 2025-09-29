package block

import (
	"github.com/df-mc/dragonfly/server/world/sound"
)

// Honeycomb is a decorative blocks crafted from honeycombs.
type Honeycomb struct {
	solid
}

func (h Honeycomb) Instrument() sound.Instrument {
	return sound.Flute()
}

func (h Honeycomb) BreakInfo() BreakInfo {
	return newBreakInfo(0.6, alwaysHarvestable, nothingEffective, oneOf(h))
}

func (Honeycomb) EncodeItem() (name string, meta int16) {
	return "minecraft:honeycomb_block", 0
}

func (Honeycomb) EncodeBlock() (string, map[string]any) {
	return "minecraft:honeycomb_block", nil
}

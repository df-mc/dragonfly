package block

import "github.com/df-mc/dragonfly/server/block/instrument"

// HoneycombBlock is a decorative blocks crafted from honeycombs.
type HoneycombBlock struct {
	solid
}

// Instrument ...
func (h HoneycombBlock) Instrument() instrument.Instrument {
	return instrument.Flute()
}

// BreakInfo ...
func (h HoneycombBlock) BreakInfo() BreakInfo {
	return newBreakInfo(0.6, alwaysHarvestable, nothingEffective, oneOf(h))
}

// EncodeItem ...
func (HoneycombBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:honeycomb_block", 0
}

// EncodeBlock ...
func (HoneycombBlock) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:honeycomb_block", nil
}

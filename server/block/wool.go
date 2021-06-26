package block

import (
	"github.com/df-mc/dragonfly/server/block/instrument"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Wool is a colourful block that can be obtained by killing/shearing sheep, or crafted using four string.
type Wool struct {
	solid

	// Colour is the colour of the wool.
	Colour item.Colour
}

// Instrument ...
func (w Wool) Instrument() instrument.Instrument {
	return instrument.Guitar()
}

// FlammabilityInfo ...
func (w Wool) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(30, 60, true)
}

// BreakInfo ...
func (w Wool) BreakInfo() BreakInfo {
	return newBreakInfo(0.8, alwaysHarvestable, shearsEffective, oneOf(w))
}

// EncodeItem ...
func (w Wool) EncodeItem() (name string, meta int16) {
	return "minecraft:wool", int16(w.Colour.Uint8())
}

// EncodeBlock ...
func (w Wool) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:wool", map[string]interface{}{"color": w.Colour.String()}
}

// allWool returns wool blocks with all possible colours.
func allWool() []world.Block {
	b := make([]world.Block, 0, 16)
	for _, c := range item.Colours() {
		b = append(b, Wool{Colour: c})
	}
	return b
}

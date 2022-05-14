package block

import (
	"github.com/df-mc/dragonfly/server/world/sound"
)

// PackedIce is an opaque solid block variant of ice. Unlike regular ice, it does not melt near bright light sources.
type PackedIce struct {
	solid
}

// Instrument ...
func (PackedIce) Instrument() sound.Instrument {
	return sound.Chimes()
}

// BreakInfo ...
func (p PackedIce) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, pickaxeEffective, silkTouchOnlyDrop(p))
}

// Friction ...
func (p PackedIce) Friction() float64 {
	return 0.98
}

// EncodeItem ...
func (PackedIce) EncodeItem() (name string, meta int16) {
	return "minecraft:packed_ice", 0
}

// EncodeBlock ...
func (PackedIce) EncodeBlock() (string, map[string]any) {
	return "minecraft:packed_ice", nil
}

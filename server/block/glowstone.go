package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world/sound"
	"math/rand"
)

// Glowstone is commonly found on the ceiling of the nether dimension.
type Glowstone struct {
	solid
}

// Instrument ...
func (g Glowstone) Instrument() sound.Instrument {
	return sound.Pling()
}

// BreakInfo ...
func (g Glowstone) BreakInfo() BreakInfo {
	return newBreakInfo(0.3, alwaysHarvestable, nothingEffective, silkTouchDrop(item.NewStack(item.GlowstoneDust{}, rand.Intn(3)+2), item.NewStack(g, 1)))
}

// EncodeItem ...
func (Glowstone) EncodeItem() (name string, meta int16) {
	return "minecraft:glowstone", 0
}

// EncodeBlock ...
func (Glowstone) EncodeBlock() (string, map[string]any) {
	return "minecraft:glowstone", nil
}

// LightEmissionLevel returns 15.
func (Glowstone) LightEmissionLevel() uint8 {
	return 15
}

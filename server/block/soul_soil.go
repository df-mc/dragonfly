package block

import "github.com/df-mc/dragonfly/server/world"

// SoulSoil is a block naturally found only in the soul sand valley.
type SoulSoil struct {
	solid
}

// SoilFor ...
func (s SoulSoil) SoilFor(block world.Block) bool {
	_, ok := block.(NetherSprouts)
	return ok
}

// BreakInfo ...
func (s SoulSoil) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, shovelEffective, oneOf(s))
}

// EncodeItem ...
func (SoulSoil) EncodeItem() (name string, meta int16) {
	return "minecraft:soul_soil", 0
}

// EncodeBlock ...
func (SoulSoil) EncodeBlock() (string, map[string]any) {
	return "minecraft:soul_soil", nil
}

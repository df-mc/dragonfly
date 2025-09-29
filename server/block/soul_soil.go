package block

import "github.com/df-mc/dragonfly/server/world"

// SoulSoil is a block naturally found only in the soul sand valley.
type SoulSoil struct {
	solid
}

func (s SoulSoil) SoilFor(block world.Block) bool {
	_, ok := block.(NetherSprouts)
	return ok
}

func (s SoulSoil) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, shovelEffective, oneOf(s))
}

func (SoulSoil) EncodeItem() (name string, meta int16) {
	return "minecraft:soul_soil", 0
}

func (SoulSoil) EncodeBlock() (string, map[string]any) {
	return "minecraft:soul_soil", nil
}

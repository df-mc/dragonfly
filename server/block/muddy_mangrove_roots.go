package block

import "github.com/df-mc/dragonfly/server/world"

// MuddyMangroveRoots are a decorative variant of mangrove roots.
type MuddyMangroveRoots struct {
	solid
}

// SoilFor ...
func (MuddyMangroveRoots) SoilFor(block world.Block) bool {
	switch block.(type) {
	case TallGrass, DoubleTallGrass, Flower, DoubleFlower, NetherSprouts:
		return true
	}
	return false
}

// BreakInfo ...
func (m MuddyMangroveRoots) BreakInfo() BreakInfo {
	return newBreakInfo(0.7, alwaysHarvestable, shovelEffective, oneOf(m))
}

// EncodeItem ...
func (MuddyMangroveRoots) EncodeItem() (name string, meta int16) {
	return "minecraft:muddy_mangrove_roots", 0
}

// EncodeBlock ...
func (MuddyMangroveRoots) EncodeBlock() (string, map[string]any) {
	return "minecraft:muddy_mangrove_roots", nil
}

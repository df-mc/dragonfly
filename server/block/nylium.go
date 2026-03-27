package block

import "github.com/df-mc/dragonfly/server/world"

// Nylium is a fungal grass-like block found in the Nether.
type Nylium struct {
	solid
	bassDrum

	// Warped specifies if the nylium is the warped variant. If false, crimson nylium is encoded.
	Warped bool
}

// SoilFor ...
func (n Nylium) SoilFor(b world.Block) bool {
	switch b.(type) {
	case NetherSprouts, Roots, Fungus:
		return true
	default:
		return false
	}
}

// BreakInfo ...
func (n Nylium) BreakInfo() BreakInfo {
	return newBreakInfo(0.4, pickaxeHarvestable, pickaxeEffective, oneOf(n))
}

// EncodeItem ...
func (n Nylium) EncodeItem() (name string, meta int16) {
	if n.Warped {
		return "minecraft:warped_nylium", 0
	}
	return "minecraft:crimson_nylium", 0
}

// EncodeBlock ...
func (n Nylium) EncodeBlock() (string, map[string]any) {
	if n.Warped {
		return "minecraft:warped_nylium", nil
	}
	return "minecraft:crimson_nylium", nil
}

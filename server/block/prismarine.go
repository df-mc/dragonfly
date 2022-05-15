package block

import (
	"github.com/df-mc/dragonfly/server/world"
)

// Prismarine is a type of stone that appears underwater in ruins and ocean monuments.
type Prismarine struct {
	solid
	bassDrum

	// Type is the type of prismarine of the block.
	Type PrismarineType
}

// BreakInfo ...
func (p Prismarine) BreakInfo() BreakInfo {
	return newBreakInfo(1.5, pickaxeHarvestable, pickaxeEffective, oneOf(p))
}

// EncodeItem ...
func (p Prismarine) EncodeItem() (id string, meta int16) {
	return "minecraft:prismarine", int16(p.Type.Uint8())
}

// EncodeBlock ...
func (p Prismarine) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:prismarine", map[string]any{"prismarine_block_type": p.Type.String()}
}

// allPrismarine returns a list of all prismarine block variants.
func allPrismarine() (c []world.Block) {
	for _, t := range PrismarineTypes() {
		c = append(c, Prismarine{Type: t})
	}

	return
}

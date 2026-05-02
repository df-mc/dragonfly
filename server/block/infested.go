package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// Infested is a block that hides a silverfish. It looks identical to its non-infested counterpart, but is broken
// twice as fast.
type Infested struct {
	solid
	bassDrum

	// Type is the type of infested block.
	Type InfestedType
}

// BreakInfo ...
func (i Infested) BreakInfo() BreakInfo {
	// Infested blocks have half the hardness of their normal counterparts. In vanilla, they drop nothing
	// and spawn a silverfish, but since mobs aren't implemented yet, we just drop nothing.
	return newBreakInfo(0.75, alwaysHarvestable, nothingEffective, nil).withBlastResistance(0.75)
}

// EncodeItem ...
func (i Infested) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_" + i.Type.String(), 0
}

// EncodeBlock ...
func (i Infested) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_" + i.Type.String(), nil
}

// allInfested ...
func allInfested() (s []world.Block) {
	for _, t := range InfestedTypes() {
		s = append(s, Infested{Type: t})
	}
	return
}

package block

import (
	"github.com/df-mc/dragonfly/server/world"
)

// InfestedStoneBricks is a block that hides a silverfish. It looks identical to stone bricks.
type InfestedStoneBricks struct {
	solid
	bassDrum
	// Type is the type of stone bricks of the block.
	Type StoneBricksType
}

// BreakInfo ...
func (s InfestedStoneBricks) BreakInfo() BreakInfo {
	return newBreakInfo(0.75, alwaysHarvestable, nothingEffective, nil).withBlastResistance(0.75)
}

// EncodeItem ...
func (s InfestedStoneBricks) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_" + s.Type.String(), 0
}

// EncodeBlock ...
func (s InfestedStoneBricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_" + s.Type.String(), nil
}

// Hash ...
func (s InfestedStoneBricks) Hash() (uint64, uint64) {
	return hashInfestedStoneBricks, uint64(s.Type.Uint8())
}

// allInfestedStoneBricks returns a list of all infested stone bricks variants.
func allInfestedStoneBricks() (s []world.Block) {
	for _, t := range StoneBricksTypes() {
		s = append(s, InfestedStoneBricks{Type: t})
	}
	return
}

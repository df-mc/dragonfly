package block

import (
	"github.com/df-mc/dragonfly/server/world"
)

// InfestedStoneBricks is a block that hides a silverfish. It looks identical to stone bricks.
type InfestedStoneBricks struct {
	solid
	flute

	// Type is the type of stone bricks of the block.
	Type StoneBricksType
}

// BreakInfo ...
func (i InfestedStoneBricks) BreakInfo() BreakInfo {
	return newBreakInfo(0.75, pickaxeHarvestable, pickaxeEffective, silkTouchOnlyDrop(i)).withBlastResistance(0.75)
}

// EncodeItem ...
func (i InfestedStoneBricks) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_" + i.Type.String(), 0
}

// EncodeBlock ...
func (i InfestedStoneBricks) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_" + i.Type.String(), nil
}

// allInfestedStoneBricks returns a list of all infested stone bricks variants.
func allInfestedStoneBricks() (s []world.Block) {
	for _, t := range StoneBricksTypes() {
		s = append(s, InfestedStoneBricks{Type: t})
	}
	return
}

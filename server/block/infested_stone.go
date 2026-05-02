package block

import (
	"github.com/df-mc/dragonfly/server/world"
)

// InfestedStone is a block that hides a silverfish. It looks identical to stone.
type InfestedStone struct {
	solid
	bassDrum
}

// BreakInfo ...
func (i InfestedStone) BreakInfo() BreakInfo {
	return newBreakInfo(0.75, alwaysHarvestable, nothingEffective, nil).withBlastResistance(0.75)
}

// EncodeItem ...
func (i InfestedStone) EncodeItem() (name string, meta int16) {
	return "minecraft:infested_stone", 0
}

// EncodeBlock ...
func (i InfestedStone) EncodeBlock() (string, map[string]any) {
	return "minecraft:infested_stone", nil
}

// Hash ...
func (i InfestedStone) Hash() (uint64, uint64) {
	return 2000, 0 // Hash temporal
}

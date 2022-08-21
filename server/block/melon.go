package block

import (
	"github.com/df-mc/dragonfly/server/item"
	"math/rand"
)

// Melon is a fruit block that grows from melon stems.
type Melon struct {
	solid
}

// BreakInfo ...
func (m Melon) BreakInfo() BreakInfo {
	return newBreakInfo(1, alwaysHarvestable, axeEffective, silkTouchDrop(item.NewStack(item.MelonSlice{}, rand.Intn(5)+3), item.NewStack(m, 1)))
}

// CompostChance ...
func (Melon) CompostChance() float64 {
	return 0.65
}

// EncodeItem ...
func (Melon) EncodeItem() (name string, meta int16) {
	return "minecraft:melon_block", 0
}

// EncodeBlock ...
func (Melon) EncodeBlock() (string, map[string]any) {
	return "minecraft:melon_block", nil
}

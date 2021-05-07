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
	// TODO: Silk touch.
	return newBreakInfo(1, alwaysHarvestable, axeEffective, simpleDrops(item.NewStack(item.MelonSlice{}, rand.Intn(5)+3)))
}

// EncodeItem ...
func (Melon) EncodeItem() (name string, meta int16) {
	return "minecraft:melon_block", 0
}

// EncodeBlock ...
func (Melon) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:melon_block", nil
}

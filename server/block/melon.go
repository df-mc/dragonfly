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
	return BreakInfo{
		Hardness:    1,
		Harvestable: alwaysHarvestable,
		Effective:   axeEffective,
		Drops:       simpleDrops(item.NewStack(item.MelonSlice{}, rand.Intn(5)+3)), //TODO: Silk Touch
	}
}

// EncodeItem ...
func (Melon) EncodeItem() (id int32, name string, meta int16) {
	return 103, "minecraft:melon_block", 0
}

// EncodeBlock ...
func (Melon) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:melon_block", nil
}

package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// CraftingTable is a utility block that allows the player to craft a variety of blocks and items.
type CraftingTable struct {
	bass
	solid
}

// EncodeItem ...
func (c CraftingTable) EncodeItem() (name string, meta int16) {
	return "minecraft:crafting_table", 0
}

// EncodeBlock ...
func (c CraftingTable) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:crafting_table", nil
}

// BreakInfo ...
func (c CraftingTable) BreakInfo() BreakInfo {
	return newBreakInfo(2.5, alwaysHarvestable, axeEffective, oneOf(c))
}

// Activate ...
func (c CraftingTable) Activate(pos cube.Pos, _ cube.Face, _ *world.World, u item.User) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos)
		return true
	}
	return false
}

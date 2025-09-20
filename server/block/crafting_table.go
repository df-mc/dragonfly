package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"time"
)

// CraftingTable is a utility block that allows the player to craft a variety of blocks and items.
type CraftingTable struct {
	bass
	solid
}

func (c CraftingTable) EncodeItem() (name string, meta int16) {
	return "minecraft:crafting_table", 0
}

func (c CraftingTable) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:crafting_table", nil
}

func (c CraftingTable) BreakInfo() BreakInfo {
	return newBreakInfo(2.5, alwaysHarvestable, axeEffective, oneOf(c))
}

func (CraftingTable) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

func (c CraftingTable) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos, tx)
		return true
	}
	return false
}

package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
)

// SmithingTable is a toolsmith's job site block that generates in villages. It can be used to upgrade diamond gear into
// netherite gear.
type SmithingTable struct {
	bass
	solid
}

// EncodeItem ...
func (SmithingTable) EncodeItem() (name string, meta int16) {
	return "minecraft:smithing_table", 0
}

// EncodeBlock ...
func (SmithingTable) EncodeBlock() (name string, properties map[string]interface{}) {
	return "minecraft:smithing_table", nil
}

// BreakInfo ...
func (s SmithingTable) BreakInfo() BreakInfo {
	return newBreakInfo(2.5, alwaysHarvestable, axeEffective, oneOf(s))
}

// Activate ...
func (SmithingTable) Activate(pos cube.Pos, _ cube.Face, _ *world.World, u item.User, _ *item.UseContext) bool {
	if opener, ok := u.(ContainerOpener); ok {
		opener.OpenBlockContainer(pos)
		return true
	}
	return false
}

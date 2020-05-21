package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block/wood"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/item/tool"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
)

// Leaves are blocks that grow as part of trees which mainly drop saplings and sticks.
type Leaves struct {
	// Wood is the type of wood of the leaves. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
	// Persistent specifies if the leaves are persistent, meaning they will not decay as a result of no wood
	// being nearby.
	Persistent bool

	shouldUpdate bool
}

// BreakInfo ...
func (l Leaves) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.2,
		Harvestable: alwaysHarvestable,
		Effective: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypeShears || t.ToolType() == tool.TypeHoe
		},
		// TODO: Add saplings and apples and drop them here.
		Drops: simpleDrops(),
	}
}

// EncodeItem ...
func (l Leaves) EncodeItem() (id int32, meta int16) {
	switch l.Wood {
	case wood.Oak():
		return 18, 0
	case wood.Spruce():
		return 18, 1
	case wood.Birch():
		return 18, 2
	case wood.Jungle():
		return 18, 3
	case wood.Acacia():
		return 161, 0
	case wood.DarkOak():
		return 161, 1
	}
	panic("invalid wood type")
}

// LightDiffusionLevel ...
func (Leaves) LightDiffusionLevel() uint8 {
	return 1
}

// EncodeBlock ...
func (l Leaves) EncodeBlock() (name string, properties map[string]interface{}) {
	switch l.Wood {
	case wood.Oak(), wood.Spruce(), wood.Birch(), wood.Jungle():
		return "minecraft:leaves", map[string]interface{}{"old_leaf_type": l.Wood.String(), "persistent_bit": l.Persistent, "update_bit": l.shouldUpdate}
	case wood.Acacia(), wood.DarkOak():
		return "minecraft:leaves2", map[string]interface{}{"new_leaf_type": l.Wood.String(), "persistent_bit": l.Persistent, "update_bit": l.shouldUpdate}
	}
	panic("invalid wood type")
}

// allLogs returns a list of all possible leaves states.
func allLeaves() (leaves []world.Block) {
	f := func(persistent, update bool) {
		leaves = append(leaves, Leaves{Wood: wood.Oak(), Persistent: persistent, shouldUpdate: update})
		leaves = append(leaves, Leaves{Wood: wood.Spruce(), Persistent: persistent, shouldUpdate: update})
		leaves = append(leaves, Leaves{Wood: wood.Birch(), Persistent: persistent, shouldUpdate: update})
		leaves = append(leaves, Leaves{Wood: wood.Jungle(), Persistent: persistent, shouldUpdate: update})
		leaves = append(leaves, Leaves{Wood: wood.Acacia(), Persistent: persistent, shouldUpdate: update})
		leaves = append(leaves, Leaves{Wood: wood.DarkOak(), Persistent: persistent, shouldUpdate: update})
	}
	f(true, true)
	f(true, false)
	f(false, true)
	f(false, false)
	return
}

package block

import (
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block/material"
	"git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/world"
)

// Leaves are blocks that grow as part of trees which mainly drop saplings and sticks.
type Leaves struct {
	// Wood is the type of wood of the leaves. This field must have one of the values found in the material
	// package. Using Leaves without a Wood type will panic.
	Wood material.Wood
	// Persistent specifies if the leaves are persistent, meaning they will not decay as a result of no wood
	// being nearby.
	Persistent bool

	shouldUpdate bool
}

// EncodeItem ...
func (l Leaves) EncodeItem() (id int32, meta int16) {
	switch l.Wood {
	case material.OakWood():
		return 18, 0
	case material.SpruceWood():
		return 18, 1
	case material.BirchWood():
		return 18, 2
	case material.JungleWood():
		return 18, 3
	case material.AcaciaWood():
		return 161, 0
	case material.DarkOakWood():
		return 161, 1
	}
	panic("invalid wood type")
}

// EncodeBlock ...
func (l Leaves) EncodeBlock() (name string, properties map[string]interface{}) {
	switch l.Wood {
	case material.OakWood(), material.SpruceWood(), material.BirchWood(), material.JungleWood():
		return "minecraft:leaves", map[string]interface{}{"old_leaf_type": l.Wood.Minecraft(), "persistent_bit": l.Persistent, "update_bit": l.shouldUpdate}
	case material.AcaciaWood(), material.DarkOakWood():
		return "minecraft:leaves2", map[string]interface{}{"new_leaf_type": l.Wood.Minecraft(), "persistent_bit": l.Persistent, "update_bit": l.shouldUpdate}
	}
	panic("invalid wood type")
}

// allLogs returns a list of all possible leaves states.
func allLeaves() (leaves []world.Block) {
	f := func(persistent, update bool) {
		leaves = append(leaves, Leaves{Wood: material.OakWood(), Persistent: persistent, shouldUpdate: update})
		leaves = append(leaves, Leaves{Wood: material.SpruceWood(), Persistent: persistent, shouldUpdate: update})
		leaves = append(leaves, Leaves{Wood: material.BirchWood(), Persistent: persistent, shouldUpdate: update})
		leaves = append(leaves, Leaves{Wood: material.JungleWood(), Persistent: persistent, shouldUpdate: update})
		leaves = append(leaves, Leaves{Wood: material.AcaciaWood(), Persistent: persistent, shouldUpdate: update})
		leaves = append(leaves, Leaves{Wood: material.DarkOakWood(), Persistent: persistent, shouldUpdate: update})
	}
	f(true, true)
	f(true, false)
	f(false, true)
	return
}

package block

import "git.jetbrains.space/dragonfly/dragonfly.git/dragonfly/block/material"

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

// Name ...
func (l Leaves) Name() string {
	if l.Wood == nil {
		panic("leaves has no wood type")
	}
	return l.Wood.Name() + " Leaves"
}

// Minecraft ...
func (l Leaves) Minecraft() (name string, properties map[string]interface{}) {
	switch l.Wood {
	case material.OakWood(), material.SpruceWood(), material.BirchWood(), material.JungleWood():
		return "minecraft:leaves", map[string]interface{}{"old_leaf_type": l.Wood.Minecraft(), "persistent_bit": l.Persistent, "update_bit": l.shouldUpdate}
	case material.AcaciaWood(), material.DarkOakWood():
		return "minecraft:leaves2", map[string]interface{}{"new_leaf_type": l.Wood.Minecraft(), "persistent_bit": l.Persistent, "update_bit": l.shouldUpdate}
	}
	panic("invalid wood type")
}

// allLogs returns a list of all possible leaves states.
func allLeaves() (leaves []Block) {
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

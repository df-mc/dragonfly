package block

import (
	"github.com/df-mc/dragonfly/dragonfly/block/wood"
	"github.com/df-mc/dragonfly/dragonfly/item"
	"github.com/df-mc/dragonfly/dragonfly/item/tool"
	"github.com/df-mc/dragonfly/dragonfly/world"
	"math/rand"
)

// Leaves are blocks that grow as part of trees which mainly drop saplings and sticks.
type Leaves struct {
	noNBT
	leaves

	// Wood is the type of wood of the leaves. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
	// Persistent specifies if the leaves are persistent, meaning they will not decay as a result of no wood
	// being nearby.
	Persistent bool

	shouldUpdate bool
}

// FlammabilityInfo ...
func (l Leaves) FlammabilityInfo() FlammabilityInfo {
	return FlammabilityInfo{
		Encouragement: 30,
		Flammability:  60,
		LavaFlammable: true,
	}
}

// BreakInfo ...
func (l Leaves) BreakInfo() BreakInfo {
	return BreakInfo{
		Hardness:    0.2,
		Harvestable: alwaysHarvestable,
		Effective: func(t tool.Tool) bool {
			return t.ToolType() == tool.TypeShears || t.ToolType() == tool.TypeHoe
		},
		Drops: func(t tool.Tool) (drops []item.Stack) {
			if t.ToolType() == tool.TypeShears { // TODO: Silk Touch
				drops = append(drops, item.NewStack(l, 1))
			} else {
				// TODO: Saplings and sticks can drop
				if (l.Wood == wood.Oak() || l.Wood == wood.DarkOak()) && rand.Float64() < 0.005 {
					drops = append(drops, item.NewStack(item.Apple{}, 1))
				}
			}
			return
		},
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

// CanDisplace ...
func (Leaves) CanDisplace(b world.Liquid) bool {
	_, ok := b.(Water)
	return ok
}

// SideClosed ...
func (Leaves) SideClosed(world.BlockPos, world.BlockPos, *world.World) bool {
	return false
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

// Hash ...
func (l Leaves) Hash() uint64 {
	return hashLeaves | (uint64(boolByte(l.Persistent)) << 32) | (uint64(boolByte(l.shouldUpdate)) << 33) | (uint64(l.Wood.Uint8()) << 34)
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

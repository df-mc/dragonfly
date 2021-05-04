package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/wood"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/item/tool"
	"github.com/df-mc/dragonfly/server/world"
	"math/rand"
)

// Leaves are blocks that grow as part of trees which mainly drop saplings and sticks.
type Leaves struct {
	leaves

	// Wood is the type of wood of the leaves. This field must have one of the values found in the material
	// package.
	Wood wood.Wood
	// Persistent specifies if the leaves are persistent, meaning they will not decay as a result of no wood
	// being nearby.
	Persistent bool

	shouldUpdate bool
}

// findLog ...
func findLog(pos cube.Pos, w *world.World, visited *[]cube.Pos, distance int) bool {
	for _, v := range *visited {
		if v == pos {
			return false
		}
	}
	*visited = append(*visited, pos)

	if log, ok := w.Block(pos).(Log); ok && !log.Stripped {
		return true
	}
	if _, ok := w.Block(pos).(Leaves); !ok || distance > 6 {
		return false
	}
	logFound := false
	pos.Neighbours(func(neighbour cube.Pos) {
		if !logFound && findLog(neighbour, w, visited, distance+1) {
			logFound = true
		}
	})
	return logFound
}

// RandomTick ...
func (l Leaves) RandomTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	if !l.Persistent && l.shouldUpdate {
		if findLog(pos, w, &[]cube.Pos{}, 0) {
			l.shouldUpdate = false
			w.PlaceBlock(pos, l)
		} else {
			w.BreakBlockWithoutParticles(pos)
		}
	}
}

// NeighbourUpdateTick ...
func (l Leaves) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if !l.Persistent && !l.shouldUpdate {
		l.shouldUpdate = true
		w.PlaceBlock(pos, l)
	}
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
		Drops: func(t tool.Tool) []item.Stack {
			if t.ToolType() == tool.TypeShears { // TODO: Silk Touch
				return []item.Stack{item.NewStack(l, 1)}
			}
			var drops []item.Stack
			if (l.Wood == wood.Oak() || l.Wood == wood.DarkOak()) && rand.Float64() < 0.005 {
				drops = append(drops, item.NewStack(item.Apple{}, 1))
			}
			// TODO: Saplings and sticks can drop along with apples
			return drops
		},
	}
}

// EncodeItem ...
func (l Leaves) EncodeItem() (name string, meta int16) {
	switch l.Wood {
	case wood.Oak(), wood.Spruce(), wood.Birch(), wood.Jungle():
		return "minecraft:leaves", int16(l.Wood.Uint8())
	case wood.Acacia(), wood.DarkOak():
		return "minecraft:leaves2", int16(l.Wood.Uint8() - 4)
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
func (Leaves) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
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

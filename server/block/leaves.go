package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/event"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
)

// Leaves are blocks that grow as part of trees which mainly drop saplings and sticks.
type Leaves struct {
	leaves
	sourceWaterDisplacer

	// Wood is the type of wood of the leaves. This field must have one of the values found in the material
	// package.
	Wood WoodType
	// Persistent specifies if the leaves are persistent, meaning they will not decay as a result of no wood
	// being nearby.
	Persistent bool

	ShouldUpdate bool
}

// UseOnBlock makes leaves persistent when they are placed so that they don't decay.
func (l Leaves) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, l)
	if !used {
		return
	}
	l.Persistent = true

	place(tx, pos, l, user, ctx)
	return placed(ctx)
}

// findLog ...
func findLog(pos cube.Pos, tx *world.Tx, visited *[]cube.Pos, distance int) bool {
	for _, v := range *visited {
		if v == pos {
			return false
		}
	}
	*visited = append(*visited, pos)

	if log, ok := tx.Block(pos).(Log); ok && !log.Stripped {
		return true
	}
	if _, ok := tx.Block(pos).(Leaves); !ok || distance > 6 {
		return false
	}
	logFound := false
	pos.Neighbours(func(neighbour cube.Pos) {
		if !logFound && findLog(neighbour, tx, visited, distance+1) {
			logFound = true
		}
	}, tx.Range())
	return logFound
}

// RandomTick ...
func (l Leaves) RandomTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if !l.Persistent && l.ShouldUpdate {
		if findLog(pos, tx, &[]cube.Pos{}, 0) {
			l.ShouldUpdate = false
			tx.SetBlock(pos, l, nil)
			return
		}
		ctx := event.C(tx)
		if tx.World().Handler().HandleLeavesDecay(ctx, pos); ctx.Cancelled() {
			// Prevent immediate re-updating.
			l.ShouldUpdate = false
			tx.SetBlock(pos, l, nil)
			return
		}
		tx.SetBlock(pos, nil, nil)
		for _, drop := range l.BreakInfo().Drops(item.ToolNone{}, nil) {
			dropItem(tx, drop, pos.Vec3Centre())
		}
	}
}

// NeighbourUpdateTick ...
func (l Leaves) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !l.Persistent && !l.ShouldUpdate {
		l.ShouldUpdate = true
		tx.SetBlock(pos, l, nil)
	}
}

// FlammabilityInfo ...
func (l Leaves) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(30, 60, true)
}

// BreakInfo ...
func (l Leaves) BreakInfo() BreakInfo {
	return newBreakInfo(0.2, alwaysHarvestable, func(t item.Tool) bool {
		return t.ToolType() == item.TypeShears || t.ToolType() == item.TypeHoe
	}, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if t.ToolType() == item.TypeShears || hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(l, 1)}
		}
		var drops []item.Stack
		// TODO: Drop saplings.
		if rand.Float64() < 0.02 {
			drops = append(drops, item.NewStack(item.Stick{}, rand.IntN(2)+1))
		}
		if (l.Wood == OakWood() || l.Wood == DarkOakWood()) && rand.Float64() < 0.005 {
			drops = append(drops, item.NewStack(item.Apple{}, 1))
		}
		return drops
	})
}

// CompostChance ...
func (Leaves) CompostChance() float64 {
	return 0.3
}

// EncodeItem ...
func (l Leaves) EncodeItem() (name string, meta int16) {
	return "minecraft:" + l.Wood.String() + "_leaves", 0
}

// LightDiffusionLevel ...
func (Leaves) LightDiffusionLevel() uint8 {
	return 1
}

// SideClosed ...
func (Leaves) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// PistonBreakable ...
func (Leaves) PistonBreakable() bool {
	return true
}

// EncodeBlock ...
func (l Leaves) EncodeBlock() (name string, properties map[string]any) {
	return "minecraft:" + l.Wood.String() + "_leaves", map[string]any{"persistent_bit": l.Persistent, "update_bit": l.ShouldUpdate}
}

// allLogs returns a list of all possible leaves states.
func allLeaves() (leaves []world.Block) {
	f := func(persistent, update bool) {
		for _, w := range WoodTypes() {
			if w != CrimsonWood() && w != WarpedWood() {
				leaves = append(leaves, Leaves{Wood: w, Persistent: persistent, ShouldUpdate: update})
			}
		}
	}
	f(true, true)
	f(true, false)
	f(false, true)
	f(false, false)
	return
}

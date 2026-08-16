package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// HangingRoots is a decorative block that hangs from the underside of blocks in lush caves.
type HangingRoots struct {
	empty
	replaceable
	transparent
	sourceWaterDisplacer
}

// NeighbourUpdateTick ...
func (h HangingRoots) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if up := pos.Side(cube.FaceUp); !tx.Block(up).Model().FaceSolid(up, cube.FaceDown, tx) {
		breakBlock(h, pos, tx)
	}
}

// UseOnBlock ...
func (h HangingRoots) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, h)
	if !used {
		return false
	}
	if up := pos.Side(cube.FaceUp); !tx.Block(up).Model().FaceSolid(up, cube.FaceDown, tx) {
		return false
	}

	place(tx, pos, h, user, ctx)
	return placed(ctx)
}

// SideClosed ...
func (HangingRoots) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// HasLiquidDrops ...
func (HangingRoots) HasLiquidDrops() bool {
	return true
}

// FlammabilityInfo ...
func (HangingRoots) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(15, 100, true)
}

// CompostChance ...
func (HangingRoots) CompostChance() float64 {
	return 0.3
}

// BreakInfo ...
func (h HangingRoots) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, shearsEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if t.ToolType() == item.TypeShears || hasSilkTouch(enchantments) {
			return []item.Stack{item.NewStack(h, 1)}
		}
		return nil
	})
}

// EncodeItem ...
func (HangingRoots) EncodeItem() (name string, meta int16) {
	return "minecraft:hanging_roots", 0
}

// EncodeBlock ...
func (HangingRoots) EncodeBlock() (string, map[string]any) {
	return "minecraft:hanging_roots", nil
}

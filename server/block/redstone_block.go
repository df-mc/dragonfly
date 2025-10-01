package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type RedstoneBlock struct {
	solid
}

// BreakInfo ...
func (r RedstoneBlock) BreakInfo() BreakInfo {
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(r)).withBlastResistance(30).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		updateAroundRedstone(pos, tx)
	})
}

// EncodeItem ...
func (r RedstoneBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_block", 0
}

// EncodeBlock ...
func (r RedstoneBlock) EncodeBlock() (string, map[string]any) {
	return "minecraft:redstone_block", nil
}

// UseOnBlock ...
func (r RedstoneBlock) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, r)
	if !used {
		return false
	}
	place(tx, pos, r, user, ctx)
	if placed(ctx) {
		updateAroundRedstone(pos, tx)
		return true
	}
	return false
}

// RedstoneSource ...
func (r RedstoneBlock) RedstoneSource() bool {
	return true
}

// WeakPower ...
func (r RedstoneBlock) WeakPower(_ cube.Pos, _ cube.Face, _ *world.Tx, _ bool) int {
	return 15
}

// StrongPower ...
func (r RedstoneBlock) StrongPower(_ cube.Pos, _ cube.Face, _ *world.Tx, _ bool) int {
	return 0
}

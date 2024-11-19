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
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(r)).withBreakHandler(func(pos cube.Pos, w *world.World, _ item.User) {
		updateAroundRedstone(pos, w)
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
func (r RedstoneBlock) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, r)
	if !used {
		return false
	}
	place(w, pos, r, user, ctx)
	if placed(ctx) {
		updateAroundRedstone(pos, w)
		return true
	}
	return false
}

// Source ...
func (r RedstoneBlock) Source() bool {
	return true
}

// WeakPower ...
func (r RedstoneBlock) WeakPower(cube.Pos, cube.Face, *world.World, bool) int {
	return 15
}

// StrongPower ...
func (r RedstoneBlock) StrongPower(cube.Pos, cube.Face, *world.World, bool) int {
	return 0
}

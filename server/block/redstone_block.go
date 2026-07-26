package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// RedstoneBlock is a mineral block equivalent to nine redstone dust.
// It acts as a permanently powered redstone power source that can be pushed by pistons.
type RedstoneBlock struct {
	solid
}

func (r RedstoneBlock) BreakInfo() BreakInfo {
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(r)).withBlastResistance(6)
}

func (r RedstoneBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_block", 0
}

func (r RedstoneBlock) EncodeBlock() (string, map[string]any) {
	return "minecraft:redstone_block", nil
}

func (r RedstoneBlock) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, r)
	if !used {
		return false
	}
	place(tx, pos, r, user, ctx)
	return placed(ctx)
}

func (RedstoneBlock) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	return 15
}

// RedstoneStrongPower returns no strong power. Redstone blocks power adjacent components directly, but do not strongly
// power adjacent opaque blocks.
func (RedstoneBlock) RedstoneStrongPower(cube.Pos, *world.Tx, cube.Face) int {
	return 0
}

func (RedstoneBlock) RedstoneNonConductive() {}

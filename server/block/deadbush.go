package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
)

// DeadBush is a transparent block in the form of an aesthetic plant.
type DeadBush struct {
	empty
	replaceable
	transparent
	sourceWaterDisplacer
}

func (d DeadBush) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsVegetation(d, tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(d, pos, tx)
	}
}

func (d DeadBush) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, d)
	if !used || !supportsVegetation(d, tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}

	place(tx, pos, d, user, ctx)
	return placed(ctx)
}

func (d DeadBush) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

func (d DeadBush) HasLiquidDrops() bool {
	return true
}

func (d DeadBush) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, true)
}

func (d DeadBush) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, func(t item.Tool, enchantments []item.Enchantment) []item.Stack {
		if t.ToolType() == item.TypeShears {
			return []item.Stack{item.NewStack(d, 1)}
		}
		if amount := rand.IntN(3); amount != 0 {
			return []item.Stack{item.NewStack(item.Stick{}, amount)}
		}
		return nil
	})
}

func (d DeadBush) EncodeItem() (name string, meta int16) {
	return "minecraft:deadbush", 0
}

func (d DeadBush) EncodeBlock() (string, map[string]any) {
	return "minecraft:deadbush", nil
}

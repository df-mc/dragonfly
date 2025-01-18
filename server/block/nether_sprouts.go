package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// NetherSprouts are a non-solid plant block that generate in warped forests.
type NetherSprouts struct {
	transparent
	replaceable
	empty
}

// NeighbourUpdateTick ...
func (n NetherSprouts) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsVegetation(n, tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(n, pos, tx) // TODO: Nylium & mycelium
	}
}

// UseOnBlock ...
func (n NetherSprouts) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, n)
	if !used {
		return false
	}
	if !supportsVegetation(n, tx.Block(pos.Side(cube.FaceDown))) {
		return false // TODO: Nylium & mycelium
	}

	place(tx, pos, n, user, ctx)
	return placed(ctx)
}

// HasLiquidDrops ...
func (n NetherSprouts) HasLiquidDrops() bool {
	return false
}

// FlammabilityInfo ...
func (n NetherSprouts) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(0, 0, true)
}

// BreakInfo ...
func (n NetherSprouts) BreakInfo() BreakInfo {
	return newBreakInfo(0, func(t item.Tool) bool {
		return t.ToolType() == item.TypeShears
	}, nothingEffective, oneOf(n))
}

// CompostChance ...
func (NetherSprouts) CompostChance() float64 {
	return 0.5
}

// EncodeItem ...
func (n NetherSprouts) EncodeItem() (name string, meta int16) {
	return "minecraft:nether_sprouts", 0
}

// EncodeBlock ...
func (n NetherSprouts) EncodeBlock() (string, map[string]any) {
	return "minecraft:nether_sprouts", nil
}

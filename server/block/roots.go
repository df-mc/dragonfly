package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Roots are non-solid Nether plants found on nylium and soul soil.
type Roots struct {
	transparent
	replaceable
	empty

	// Warped specifies if the roots are the warped variant. If false, crimson roots are encoded.
	Warped bool
}

// NeighbourUpdateTick ...
func (r Roots) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsNetherRoots(tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(r, pos, tx)
	}
}

// UseOnBlock ...
func (r Roots) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, r)
	if !used || !supportsNetherRoots(tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}
	place(tx, pos, r, user, ctx)
	return placed(ctx)
}

// HasLiquidDrops ...
func (Roots) HasLiquidDrops() bool {
	return false
}

// FlammabilityInfo ...
func (Roots) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, true)
}

// BreakInfo ...
func (r Roots) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(r))
}

// CompostChance ...
func (Roots) CompostChance() float64 {
	return 0.65
}

// EncodeItem ...
func (r Roots) EncodeItem() (name string, meta int16) {
	if r.Warped {
		return "minecraft:warped_roots", 0
	}
	return "minecraft:crimson_roots", 0
}

// EncodeBlock ...
func (r Roots) EncodeBlock() (string, map[string]any) {
	if r.Warped {
		return "minecraft:warped_roots", nil
	}
	return "minecraft:crimson_roots", nil
}

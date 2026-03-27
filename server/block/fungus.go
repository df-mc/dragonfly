package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// Fungus is a non-solid Nether plant that grows on nylium.
type Fungus struct {
	transparent
	replaceable
	empty

	// Warped specifies if the fungus is the warped variant. If false, crimson fungus is encoded.
	Warped bool
}

// NeighbourUpdateTick ...
func (f Fungus) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !supportsNetherFlora(tx.Block(pos.Side(cube.FaceDown))) {
		breakBlock(f, pos, tx)
	}
}

// UseOnBlock ...
func (f Fungus) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, f)
	if !used || !supportsNetherFlora(tx.Block(pos.Side(cube.FaceDown))) {
		return false
	}
	place(tx, pos, f, user, ctx)
	return placed(ctx)
}

// HasLiquidDrops ...
func (Fungus) HasLiquidDrops() bool {
	return false
}

// FlammabilityInfo ...
func (Fungus) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(60, 100, true)
}

// BreakInfo ...
func (f Fungus) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(f))
}

// CompostChance ...
func (Fungus) CompostChance() float64 {
	return 0.65
}

// EncodeItem ...
func (f Fungus) EncodeItem() (name string, meta int16) {
	if f.Warped {
		return "minecraft:warped_fungus", 0
	}
	return "minecraft:crimson_fungus", 0
}

// EncodeBlock ...
func (f Fungus) EncodeBlock() (string, map[string]any) {
	if f.Warped {
		return "minecraft:warped_fungus", nil
	}
	return "minecraft:crimson_fungus", nil
}

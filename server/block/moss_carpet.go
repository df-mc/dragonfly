package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// MossCarpet is a thin decorative variant of the moss block.
type MossCarpet struct {
	carpet
	transparent
	sourceWaterDisplacer
}

// SideClosed ...
func (MossCarpet) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// HasLiquidDrops ...
func (MossCarpet) HasLiquidDrops() bool {
	return true
}

// NeighbourUpdateTick ...
func (m MossCarpet) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Air); ok {
		breakBlock(m, pos, tx)
	}
}

// UseOnBlock ...
func (m MossCarpet) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, m)
	if !used {
		return
	}
	if _, ok := tx.Block(pos.Side(cube.FaceDown)).(Air); ok {
		return
	}

	place(tx, pos, m, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (m MossCarpet) BreakInfo() BreakInfo {
	return newBreakInfo(0.1, alwaysHarvestable, nothingEffective, oneOf(m))
}

// CompostChance ...
func (MossCarpet) CompostChance() float64 {
	return 0.3
}

// EncodeItem ...
func (m MossCarpet) EncodeItem() (name string, meta int16) {
	return "minecraft:moss_carpet", 0
}

// EncodeBlock ...
func (m MossCarpet) EncodeBlock() (string, map[string]any) {
	return "minecraft:moss_carpet", nil
}

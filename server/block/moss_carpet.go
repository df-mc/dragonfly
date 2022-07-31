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
}

// CanDisplace ...
func (MossCarpet) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (MossCarpet) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// HasLiquidDrops ...
func (MossCarpet) HasLiquidDrops() bool {
	return true
}

// NeighbourUpdateTick ...
func (MossCarpet) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if _, ok := w.Block(pos.Side(cube.FaceDown)).(Air); ok {
		w.SetBlock(pos, nil, nil)
	}
}

// UseOnBlock ...
func (m MossCarpet) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, m)
	if !used {
		return
	}
	if _, ok := w.Block(pos.Side(cube.FaceDown)).(Air); ok {
		return
	}

	place(w, pos, m, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (m MossCarpet) BreakInfo() BreakInfo {
	return newBreakInfo(0.1, alwaysHarvestable, nothingEffective, oneOf(m))
}

// EncodeItem ...
func (m MossCarpet) EncodeItem() (name string, meta int16) {
	return "minecraft:moss_carpet", 0
}

// EncodeBlock ...
func (m MossCarpet) EncodeBlock() (string, map[string]any) {
	return "minecraft:moss_carpet", nil
}

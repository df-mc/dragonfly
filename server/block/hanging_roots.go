package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// HangingRoots are a natural decorative block found underground in the lush caves biome.
type HangingRoots struct {
	empty
	transparent
}

// FlammabilityInfo ...
func (h HangingRoots) FlammabilityInfo() FlammabilityInfo {
	return newFlammabilityInfo(30, 60, true)
}

// UseOnBlock ...
func (h HangingRoots) UseOnBlock(pos cube.Pos, face cube.Face, clickPos mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, h)
	if !used {
		return false
	}
	if !w.Block(pos.Side(cube.FaceUp)).Model().FaceSolid(pos.Side(cube.FaceUp), cube.FaceDown, w) {
		return false
	}

	place(w, pos, h, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (h HangingRoots) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if !w.Block(pos.Side(cube.FaceUp)).Model().FaceSolid(pos.Side(cube.FaceUp), cube.FaceDown, w) {
		w.BreakBlock(pos)
	}
}

// BreakInfo ...
func (h HangingRoots) BreakInfo() BreakInfo {
	return newBreakInfo(0.1, shearsEffective, nothingEffective, oneOf(h))
}

// EncodeItem ...
func (h HangingRoots) EncodeItem() (name string, meta int16) {
	return "minecraft:hanging_roots", 0
}

// EncodeBlock ...
func (h HangingRoots) EncodeBlock() (string, map[string]interface{}) {
	return "minecraft:hanging_roots", nil
}

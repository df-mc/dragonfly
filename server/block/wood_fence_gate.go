package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// WoodFenceGate is a block that can be used as an openable 1x1 barrier.
type WoodFenceGate struct {
	transparent
	bass

	// Wood is the type of wood of the fence gate. This field must have one of the values found in the material
	// package.
	Wood WoodType
	// Facing is the direction the fence gate swings open.
	Facing cube.Direction
	// Open is whether the fence gate is open.
	Open bool
	// Lowered lowers the fence gate by 3 pixels and is set when placed next to wall blocks.
	Lowered bool
}

// BreakInfo ...
func (f WoodFenceGate) BreakInfo() BreakInfo {
	return newBreakInfo(2, alwaysHarvestable, axeEffective, oneOf(f))
}

// FlammabilityInfo ...
func (f WoodFenceGate) FlammabilityInfo() FlammabilityInfo {
	if !f.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(5, 20, true)
}

// UseOnBlock ...
func (f WoodFenceGate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, f)
	if !used {
		return false
	}
	f.Facing = user.Facing()
	// TODO: Set Lowered if placed next to wall block.

	place(w, pos, f, user, ctx)
	return placed(ctx)
}

// Activate ...
func (f WoodFenceGate) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User) bool {
	f.Open = !f.Open
	if f.Open && f.Facing.Opposite() == u.Facing() {
		f.Facing = u.Facing()
	}
	w.SetBlock(pos, f, nil)
	w.PlaySound(pos.Vec3Centre(), sound.Door{})
	return true
}

// CanDisplace ...
func (f WoodFenceGate) CanDisplace(b world.Liquid) bool {
	_, water := b.(Water)
	return water
}

// SideClosed ...
func (f WoodFenceGate) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// EncodeItem ...
func (f WoodFenceGate) EncodeItem() (name string, meta int16) {
	if f.Wood == OakWood() {
		return "minecraft:fence_gate", 0
	}
	return "minecraft:" + f.Wood.String() + "_fence_gate", 0
}

// EncodeBlock ...
func (f WoodFenceGate) EncodeBlock() (name string, properties map[string]any) {
	if f.Wood == OakWood() {
		return "minecraft:fence_gate", map[string]any{"direction": int32(horizontalDirection(f.Facing)), "open_bit": f.Open, "in_wall_bit": f.Lowered}
	}
	return "minecraft:" + f.Wood.String() + "_fence_gate", map[string]any{"direction": int32(horizontalDirection(f.Facing)), "open_bit": f.Open, "in_wall_bit": f.Lowered}
}

// Model ...
func (f WoodFenceGate) Model() world.BlockModel {
	return model.FenceGate{Facing: f.Facing, Open: f.Open}
}

// allFenceGates returns a list of all trapdoor types.
func allFenceGates() (fenceGates []world.Block) {
	for _, w := range WoodTypes() {
		for i := cube.Direction(0); i <= 3; i++ {
			fenceGates = append(fenceGates, WoodFenceGate{Wood: w, Facing: i, Open: false, Lowered: false})
			fenceGates = append(fenceGates, WoodFenceGate{Wood: w, Facing: i, Open: false, Lowered: true})
			fenceGates = append(fenceGates, WoodFenceGate{Wood: w, Facing: i, Open: true, Lowered: true})
			fenceGates = append(fenceGates, WoodFenceGate{Wood: w, Facing: i, Open: true, Lowered: false})
		}
	}
	return
}

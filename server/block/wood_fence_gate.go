package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"time"
)

// WoodFenceGate is a block that can be used as an openable 1x1 barrier.
type WoodFenceGate struct {
	transparent
	bass
	sourceWaterDisplacer

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
	return newBreakInfo(2, alwaysHarvestable, axeEffective, oneOf(f)).withBlastResistance(15)
}

// FlammabilityInfo ...
func (f WoodFenceGate) FlammabilityInfo() FlammabilityInfo {
	if !f.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(5, 20, true)
}

// FuelInfo ...
func (WoodFenceGate) FuelInfo() item.FuelInfo {
	return newFuelInfo(time.Second * 15)
}

// UseOnBlock ...
func (f WoodFenceGate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, f)
	if !used {
		return false
	}
	f.Facing = user.Rotation().Direction()
	f.Lowered = f.shouldBeLowered(pos, w)

	place(w, pos, f, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (f WoodFenceGate) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if f.shouldBeLowered(pos, w) != f.Lowered {
		f.Lowered = !f.Lowered
		w.SetBlock(pos, f, nil)
	}
}

// shouldBeLowered returns if the fence gate should be lowered or not, based on the neighbouring walls.
func (f WoodFenceGate) shouldBeLowered(pos cube.Pos, w *world.World) bool {
	leftSide := f.Facing.RotateLeft().Face()
	_, left := w.Block(pos.Side(leftSide)).(Wall)
	_, right := w.Block(pos.Side(leftSide.Opposite())).(Wall)
	return left || right
}

// Activate ...
func (f WoodFenceGate) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User, _ *item.UseContext) bool {
	f.Open = !f.Open
	if f.Open && f.Facing.Opposite() == u.Rotation().Direction() {
		f.Facing = f.Facing.Opposite()
	}
	w.SetBlock(pos, f, nil)
	if f.Open {
		w.PlaySound(pos.Vec3Centre(), sound.FenceGateOpen{Block: f})
		return true
	}
	w.PlaySound(pos.Vec3Centre(), sound.FenceGateClose{Block: f})
	return true
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

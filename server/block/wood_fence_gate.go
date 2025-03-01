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
func (f WoodFenceGate) FuelInfo() item.FuelInfo {
	if !f.Wood.Flammable() {
		return item.FuelInfo{}
	}
	return newFuelInfo(time.Second * 15)
}

// UseOnBlock ...
func (f WoodFenceGate) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, f)
	if !used {
		return false
	}
	f.Facing = user.Rotation().Direction()
	f.Lowered = f.shouldBeLowered(pos, tx)

	place(tx, pos, f, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (f WoodFenceGate) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if f.shouldBeLowered(pos, tx) != f.Lowered {
		f.Lowered = !f.Lowered
		tx.SetBlock(pos, f, nil)
	}
}

// shouldBeLowered returns if the fence gate should be lowered or not, based on the neighbouring walls.
func (f WoodFenceGate) shouldBeLowered(pos cube.Pos, tx *world.Tx) bool {
	leftSide := f.Facing.RotateLeft().Face()
	_, left := tx.Block(pos.Side(leftSide)).(Wall)
	_, right := tx.Block(pos.Side(leftSide.Opposite())).(Wall)
	return left || right
}

// Activate ...
func (f WoodFenceGate) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, u item.User, _ *item.UseContext) bool {
	f.Open = !f.Open
	if f.Open && f.Facing.Opposite() == u.Rotation().Direction() {
		f.Facing = f.Facing.Opposite()
	}
	tx.SetBlock(pos, f, nil)
	if f.Open {
		tx.PlaySound(pos.Vec3Centre(), sound.FenceGateOpen{Block: f})
		return true
	}
	tx.PlaySound(pos.Vec3Centre(), sound.FenceGateClose{Block: f})
	return true
}

// SideClosed ...
func (f WoodFenceGate) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
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
		return "minecraft:fence_gate", map[string]any{"minecraft:cardinal_direction": f.Facing.String(), "open_bit": f.Open, "in_wall_bit": f.Lowered}
	}
	return "minecraft:" + f.Wood.String() + "_fence_gate", map[string]any{"minecraft:cardinal_direction": f.Facing.String(), "open_bit": f.Open, "in_wall_bit": f.Lowered}
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

package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
	"time"
)

// WoodButton is a non-solid block that can provide temporary redstone power when pressed.
type WoodButton struct {
	empty
	transparent
	flowingWaterDisplacer

	// Wood is the type of wood of the button. This field must have one of the values found in the material
	// package.
	Wood WoodType
	// Powered reports whether the button is currently pressed.
	Powered bool
	// Facing is the face of the block that the button is attached to.
	Facing cube.Face
}

// RedstoneSource ...
func (WoodButton) RedstoneSource() bool {
	return true
}

// WeakPower ...
func (b WoodButton) WeakPower(cube.Pos, cube.Face, *world.Tx, bool) int {
	if b.Powered {
		return 15
	}
	return 0
}

// StrongPower ...
func (b WoodButton) StrongPower(_ cube.Pos, face cube.Face, _ *world.Tx, _ bool) int {
	if b.Powered && b.Facing == face {
		return 15
	}
	return 0
}

// SideClosed ...
func (WoodButton) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// NeighbourUpdateTick ...
func (b WoodButton) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	supportPos := pos.Side(b.Facing.Opposite())
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, b.Facing, tx) {
		breakBlock(b, pos, tx)
	}
}

// UseOnBlock ...
func (b WoodButton) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, b)
	if !used {
		return false
	}
	supportPos := pos.Side(face.Opposite())
	if !tx.Block(supportPos).Model().FaceSolid(supportPos, face, tx) {
		return false
	}

	b.Powered = false
	b.Facing = face

	place(tx, pos, b, user, ctx)
	return placed(ctx)
}

// Activate ...
func (b WoodButton) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, _ item.User, _ *item.UseContext) bool {
	if b.Powered {
		return false
	}
	b.Powered = true
	tx.SetBlock(pos, b, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.Click{})
	tx.ScheduleBlockUpdate(pos, b, time.Millisecond*750)
	updateDirectionalRedstone(pos, tx, b.Facing.Opposite())
	return true
}

// ScheduledTick ...
func (b WoodButton) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if !b.Powered {
		return
	}
	b.Powered = false
	tx.SetBlock(pos, b, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.Click{})
	updateDirectionalRedstone(pos, tx, b.Facing.Opposite())
}

// BreakInfo ...
func (b WoodButton) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, axeEffective, oneOf(b)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		if b.Powered {
			updateDirectionalRedstone(pos, tx, b.Facing.Opposite())
		}
	})
}

// FlammabilityInfo ...
func (b WoodButton) FlammabilityInfo() FlammabilityInfo {
	if !b.Wood.Flammable() {
		return newFlammabilityInfo(0, 0, false)
	}
	return newFlammabilityInfo(5, 20, true)
}

// FuelInfo ...
func (b WoodButton) FuelInfo() item.FuelInfo {
	if !b.Wood.Flammable() {
		return item.FuelInfo{}
	}
	return newFuelInfo(time.Second * 10)
}

// EncodeItem ...
func (b WoodButton) EncodeItem() (name string, meta int16) {
	if b.Wood == OakWood() {
		return "minecraft:wooden_button", 0
	}
	return "minecraft:" + b.Wood.String() + "_button", 0
}

// EncodeBlock ...
func (b WoodButton) EncodeBlock() (name string, properties map[string]any) {
	facingInt := int32(b.Facing)
	if b.Wood == OakWood() {
		return "minecraft:wooden_button", map[string]any{"button_pressed_bit": b.Powered, "facing_direction": facingInt}
	}
	return "minecraft:" + b.Wood.String() + "_button", map[string]any{"button_pressed_bit": b.Powered, "facing_direction": facingInt}
}

// allWoodButtons returns a list of all wood button types.
func allWoodButtons() (buttons []world.Block) {
	for _, w := range WoodTypes() {
		for _, face := range cube.Faces() {
			buttons = append(buttons, WoodButton{Wood: w, Facing: face})
			buttons = append(buttons, WoodButton{Wood: w, Facing: face, Powered: true})
		}
	}
	return
}

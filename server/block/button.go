package block

import (
	"math/rand/v2"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
)

// Button is a non-solid block that emits redstone power for a short duration
// when pressed.
type Button struct {
	empty
	transparent
	sourceWaterDisplacer

	// Type is the material the button is made of.
	Type ButtonType
	// Facing is the face of the block that the button is attached to.
	Facing cube.Face
	// Pressed is true while the button emits power.
	Pressed bool
}

// Model ...
func (Button) Model() world.BlockModel {
	return model.Empty{}
}

// UseOnBlock places the button attached to the clicked face.
func (b Button) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, b)
	if !used || !redstoneAttachmentSupported(tx, pos, face) {
		return false
	}
	b.Facing = face
	place(tx, pos, b, user, ctx)
	return placed(ctx)
}

// Activate presses the button and schedules its release.
func (b Button) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, _ item.User, _ *item.UseContext) bool {
	b.press(pos, tx)
	return true
}

// ProjectileHit presses wooden buttons hit by an arrow.
func (b Button) ProjectileHit(pos cube.Pos, tx *world.Tx, e world.Entity, _ cube.Face) {
	if !b.Type.Wood() || e.H().Type().EncodeEntity() != "minecraft:arrow" {
		return
	}
	b.press(pos, tx)
}

// press activates an unpressed button and schedules its release.
func (b Button) press(pos cube.Pos, tx *world.Tx) {
	if b.Pressed {
		return
	}
	b.Pressed = true
	tx.SetBlock(pos, b, nil)
	tx.ScheduleBlockUpdate(pos, b, b.pressDuration())
	tx.PlaySound(pos.Vec3Centre(), sound.Click{})
}

// NeighbourUpdateTick breaks the button if its supporting block is removed.
func (b Button) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !redstoneAttachmentSupported(tx, pos, b.Facing) {
		breakBlock(b, pos, tx)
	}
}

// ScheduledTick releases a pressed button, unless an arrow rests inside a
// wooden button, keeping it pressed.
func (b Button) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if !b.Pressed {
		return
	}
	if b.Type.Wood() && arrowWithin(pos, tx) {
		tx.ScheduleBlockUpdate(pos, b, b.pressDuration())
		return
	}
	b.Pressed = false
	tx.SetBlock(pos, b, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.Click{})
}

// arrowWithin reports whether an arrow intersects the block space at pos.
func arrowWithin(pos cube.Pos, tx *world.Tx) bool {
	box := cube.Box(0, 0, 0, 1, 1, 1).Translate(pos.Vec3())
	for e := range tx.EntitiesWithin(box) {
		if e.H().Type().EncodeEntity() == "minecraft:arrow" {
			return true
		}
	}
	return false
}

// RedstonePower returns maximum power while the button is pressed.
func (b Button) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	if b.Pressed {
		return 15
	}
	return 0
}

// RedstoneStrongPower strongly powers the block the button is attached to.
func (b Button) RedstoneStrongPower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if b.Pressed && face == b.Facing.Opposite() {
		return 15
	}
	return 0
}

// BreakInfo ...
func (b Button) BreakInfo() BreakInfo {
	return newBreakInfo(0.5, alwaysHarvestable, pickaxeEffective, oneOf(Button{Type: b.Type}))
}

// SideClosed ...
func (Button) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// FuelInfo ...
func (b Button) FuelInfo() item.FuelInfo {
	if b.Type.Flammable() {
		return newFuelInfo(time.Second * 5)
	}
	return item.FuelInfo{}
}

// EncodeItem ...
func (b Button) EncodeItem() (name string, meta int16) {
	return "minecraft:" + b.Type.String(), 0
}

// EncodeBlock ...
func (b Button) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + b.Type.String(), map[string]any{"button_pressed_bit": boolByte(b.Pressed), "facing_direction": int32(b.Facing)}
}

// pressDuration returns how long the button stays pressed: 1.5 seconds for
// wooden buttons and 1 second for stone-like buttons.
func (b Button) pressDuration() time.Duration {
	if b.Type.Wood() {
		return time.Second * 3 / 2
	}
	return time.Second
}

// allButtons ...
func allButtons() (buttons []world.Block) {
	for _, t := range ButtonTypes() {
		for _, face := range cube.Faces() {
			buttons = append(buttons, Button{Type: t, Facing: face}, Button{Type: t, Facing: face, Pressed: true})
		}
	}
	return
}

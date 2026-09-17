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

// Button emits redstone power for a short time when pressed.
type Button struct {
	empty
	transparent
	sourceWaterDisplacer

	// Type is the button material.
	Type ButtonType
	// Facing is the face the button points towards.
	Facing cube.Face
	// Pressed reports whether the button is active.
	Pressed bool
}

// UseOnBlock places a button on the clicked face.
func (b Button) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(tx, pos, face, b)
	if !used || !buttonAttachmentSupported(tx, pos, face) {
		return false
	}
	b.Facing = face
	place(tx, pos, b, user, ctx)
	return placed(ctx)
}

// Activate presses the button.
func (b Button) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, _ item.User, _ *item.UseContext) bool {
	b.press(pos, tx)
	return true
}

// ProjectileHit presses a wooden button when struck by an activating projectile.
func (b Button) ProjectileHit(pos cube.Pos, tx *world.Tx, e world.Entity, _ cube.Face) {
	b.EntityInside(pos, tx, e)
}

// EntityInside presses a wooden button touched by an arrow or thrown trident.
func (b Button) EntityInside(pos cube.Pos, tx *world.Tx, e world.Entity) {
	if b.Type.Wood() && b.activatingProjectileIntersects(e, buttonActivationBox(b.Facing).Translate(pos.Vec3())) {
		b.press(pos, tx)
	}
}

func (b Button) press(pos cube.Pos, tx *world.Tx) {
	if b.Pressed {
		return
	}
	b.Pressed = true
	tx.SetBlock(pos, b, nil)
	tx.ScheduleBlockUpdate(pos, b, b.pressDuration())
	tx.PlaySound(pos.Vec3Centre(), sound.ButtonClickOn{Block: b})
}

// NeighbourUpdateTick breaks an unsupported button.
func (b Button) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !buttonAttachmentSupported(tx, pos, b.Facing) {
		breakBlock(b, pos, tx)
	}
}

// buttonAttachmentSupported checks full-face support without using stair corner geometry.
func buttonAttachmentSupported(tx *world.Tx, pos cube.Pos, face cube.Face) bool {
	support := pos.Side(face.Opposite())
	if support.OutOfBounds(tx.Range()) {
		return false
	}
	m := tx.Block(support).Model()
	switch m := m.(type) {
	case model.Fence, model.Wall, model.Thin:
		return false
	case model.Stair:
		if face.Axis() != cube.Y {
			return face == m.Facing.Face()
		}
	}
	return m.FaceSolid(support, face, tx)
}

// ScheduledTick releases the button unless an activating projectile still holds it down.
func (b Button) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if !b.Pressed {
		return
	}
	if b.Type.Wood() && b.activatingProjectileWithin(pos, tx) {
		tx.ScheduleBlockUpdate(pos, b, b.pressDuration())
		return
	}
	b.Pressed = false
	tx.SetBlock(pos, b, nil)
	tx.PlaySound(pos.Vec3Centre(), sound.ButtonClickOff{Block: b})
}

// activatingProjectileWithin reports whether an activating projectile intersects the button.
func (b Button) activatingProjectileWithin(pos cube.Pos, tx *world.Tx) bool {
	box := buttonActivationBox(b.Facing).Translate(pos.Vec3())
	for e := range tx.EntitiesWithin(box.Grow(1)) {
		if b.activatingProjectileIntersects(e, box) {
			return true
		}
	}
	return false
}

// activatingProjectileIntersects reports whether an arrow or thrown trident intersects a box.
func (Button) activatingProjectileIntersects(e world.Entity, box cube.BBox) bool {
	switch e.H().Type().EncodeEntity() {
	case "minecraft:arrow", "minecraft:thrown_trident":
		return entityIntersects(e, box)
	default:
		return false
	}
}

// buttonActivationBox returns the projectile detection box, which does not shrink when pressed.
func buttonActivationBox(face cube.Face) cube.BBox {
	long, short := cube.X, cube.Z
	switch face.Axis() {
	case cube.X:
		long, short = cube.Z, cube.Y
	case cube.Z:
		short = cube.Y
	}
	return cube.Box(0.5, 0.5, 0.5, 0.5, 0.5, 0.5).
		Stretch(long, 3.0/16).Stretch(short, 2.0/16).
		TranslateTowards(face.Opposite(), 0.5).
		ExtendTowards(face, 2.0/16)
}

// RedstonePower returns 15 while the button is pressed.
func (b Button) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	if b.Pressed {
		return 15
	}
	return 0
}

// RedstoneStrongPower powers the block behind the button.
func (b Button) RedstoneStrongPower(_ cube.Pos, _ *world.Tx, face cube.Face) int {
	if b.Pressed && face == b.Facing.Opposite() {
		return 15
	}
	return 0
}

// BreakInfo returns the button's break information.
func (b Button) BreakInfo() BreakInfo {
	effective := pickaxeEffective
	harvestable := pickaxeHarvestable
	if b.Type.Wood() {
		effective = axeEffective
		harvestable = alwaysHarvestable
	}
	return newBreakInfo(0.5, harvestable, effective, oneOf(Button{Type: b.Type}))
}

// SideClosed reports that buttons do not close block faces.
func (Button) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// FuelInfo returns the button's fuel properties.
func (b Button) FuelInfo() item.FuelInfo {
	if b.Type.Flammable() {
		return newFuelInfo(time.Second * 5)
	}
	return item.FuelInfo{}
}

// EncodeItem encodes the button as an item.
func (b Button) EncodeItem() (name string, meta int16) {
	return "minecraft:" + b.Type.String(), 0
}

// EncodeBlock encodes the button as a block.
func (b Button) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + b.Type.String(), map[string]any{"button_pressed_bit": boolByte(b.Pressed), "facing_direction": int32(b.Facing)}
}

func (b Button) pressDuration() time.Duration {
	if b.Type.Wood() {
		return time.Second * 3 / 2
	}
	return time.Second
}

func allButtons() (buttons []world.Block) {
	for _, t := range ButtonTypes() {
		for _, face := range cube.Faces() {
			buttons = append(buttons, Button{Type: t, Facing: face}, Button{Type: t, Facing: face, Pressed: true})
		}
	}
	return
}

package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/sound"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// TODO: Activate on projectile hit

// Button is a non-solid block that can provide temporary redstone power.
type Button struct {
	empty
	transparent
	sourceWaterDisplacer

	// Type is the type of the button.
	Type ButtonType
	// Facing is the face of the block that the button is on.
	Facing cube.Face
	// Pressed is whether the button is pressed or not.
	Pressed bool
}

// FuelInfo ...
func (b Button) FuelInfo() item.FuelInfo {
	if b.Type == StoneButton() || b.Type == PolishedBlackstoneButton() {
		return item.FuelInfo{}
	}
	return newFuelInfo(time.Second * 5)
}

// Source ...
func (b Button) Source() bool {
	return true
}

// WeakPower ...
func (b Button) WeakPower(cube.Pos, cube.Face, *world.World, bool) int {
	if b.Pressed {
		return 15
	}
	return 0
}

// StrongPower ...
func (b Button) StrongPower(_ cube.Pos, face cube.Face, _ *world.World, _ bool) int {
	if b.Pressed && b.Facing == face {
		return 15
	}
	return 0
}

// ScheduledTick ...
func (b Button) ScheduledTick(pos cube.Pos, w *world.World, r *rand.Rand) {
	if !b.Pressed {
		return
	}
	b.Pressed = false
	w.SetBlock(pos, b, nil)
	w.PlaySound(pos.Vec3Centre(), sound.PowerOff{})
	updateDirectionalRedstone(pos, w, b.Facing.Opposite())
}

// NeighbourUpdateTick ...
func (b Button) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if !w.Block(pos.Side(b.Facing.Opposite())).Model().FaceSolid(pos.Side(b.Facing.Opposite()), b.Facing, w) {
		w.SetBlock(pos, nil, nil)
		dropItem(w, item.NewStack(b, 1), pos.Vec3Centre())
		updateDirectionalRedstone(pos, w, b.Facing.Opposite())
	}
}

// Activate ...
func (b Button) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User, ctx *item.UseContext) bool {
	if b.Pressed {
		return true
	}
	b.Pressed = true
	w.SetBlock(pos, b, nil)
	w.PlaySound(pos.Vec3Centre(), sound.PowerOn{})
	updateDirectionalRedstone(pos, w, b.Facing.Opposite())

	delay := time.Millisecond * 1500
	if b.Type == StoneButton() || b.Type == PolishedBlackstoneButton() {
		delay = time.Millisecond * 1000
	}
	w.ScheduleBlockUpdate(pos, delay)
	return true

}

// UseOnBlock ...
func (b Button) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, face, used := firstReplaceable(w, pos, face, b)
	if !used {
		return false
	}
	if !w.Block(pos.Side(face.Opposite())).Model().FaceSolid(pos.Side(face.Opposite()), face, w) {
		return false
	}

	b.Facing = face
	place(w, pos, b, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (b Button) BreakInfo() BreakInfo {
	harvestTool := alwaysHarvestable
	effectiveTool := axeEffective
	if b.Type == StoneButton() || b.Type == PolishedBlackstoneButton() {
		harvestTool = pickaxeHarvestable
		effectiveTool = pickaxeEffective
	}
	return newBreakInfo(0.5, harvestTool, effectiveTool, oneOf(b)).withBreakHandler(func(pos cube.Pos, w *world.World, _ item.User) {
		updateDirectionalRedstone(pos, w, b.Facing.Opposite())
	})
}

// EncodeItem ...
func (b Button) EncodeItem() (name string, meta int16) {
	return "minecraft:" + b.Type.String() + "_button", 0
}

// EncodeBlock ...
func (b Button) EncodeBlock() (string, map[string]any) {
	return "minecraft:" + b.Type.String() + "_button", map[string]any{"facing_direction": int32(b.Facing), "button_pressed_bit": b.Pressed}
}

// allButtons ...
func allButtons() (buttons []world.Block) {
	for _, w := range ButtonTypes() {
		for _, f := range cube.Faces() {
			buttons = append(buttons, Button{Type: w, Facing: f})
			buttons = append(buttons, Button{Type: w, Facing: f, Pressed: true})
		}
	}
	return
}

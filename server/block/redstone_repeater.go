package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
	"time"
)

// RedstoneRepeater is a block used in redstone circuits to "repeat" redstone signals back to full strength, delay
// signals, prevent signals moving backwards, or to "lock" signals in one state.
type RedstoneRepeater struct {
	transparent

	// Facing is the direction from the torch to the block.
	Facing cube.Direction
	// Powered is true if the repeater is powered by a redstone signal.
	Powered bool
	// Delay represents the delay of the repeater in redstone ticks. It is between the range of one to four.
	Delay int
}

// HasLiquidDrops ...
func (RedstoneRepeater) HasLiquidDrops() bool {
	return true
}

// Model ...
func (RedstoneRepeater) Model() world.BlockModel {
	return model.Diode{}
}

// BreakInfo ...
func (r RedstoneRepeater) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(r)).withBreakHandler(func(pos cube.Pos, w *world.World, _ item.User) {
		updateSurroundingRedstone(pos, w)
	})
}

// EncodeItem ...
func (RedstoneRepeater) EncodeItem() (name string, meta int16) {
	return "minecraft:repeater", 0
}

// EncodeBlock ...
func (r RedstoneRepeater) EncodeBlock() (string, map[string]any) {
	if r.Powered {
		return "minecraft:powered_repeater", map[string]any{"direction": int32(horizontalDirection(r.Facing)), "repeater_delay": int32(r.Delay)}
	}
	return "minecraft:unpowered_repeater", map[string]any{"direction": int32(horizontalDirection(r.Facing)), "repeater_delay": int32(r.Delay)}
}

// UseOnBlock ...
func (r RedstoneRepeater) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, r)
	if !used {
		return false
	}
	if d, ok := w.Block(pos.Side(cube.FaceDown)).(LightDiffuser); ok && d.LightDiffusionLevel() == 0 {
		return false
	}
	r.Facing = user.Rotation().Direction().Opposite()

	place(w, pos, r, user, ctx)
	if placed(ctx) {
		r.RedstoneUpdate(pos, w)
		return true
	}
	return false
}

// NeighbourUpdateTick ...
func (r RedstoneRepeater) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	if d, ok := w.Block(pos.Side(cube.FaceDown)).(LightDiffuser); ok && d.LightDiffusionLevel() == 0 {
		w.SetBlock(pos, nil, nil)
		dropItem(w, item.NewStack(r, 1), pos.Vec3Centre())
		return
	}
	r.RedstoneUpdate(pos, w)
}

// Activate ...
func (r RedstoneRepeater) Activate(pos cube.Pos, _ cube.Face, w *world.World, u item.User, _ *item.UseContext) bool {
	if r.Delay++; r.Delay > 3 {
		r.Delay = 0
	}
	w.SetBlock(pos, r, nil)
	return false
}

// RedstoneUpdate ...
func (r RedstoneRepeater) RedstoneUpdate(pos cube.Pos, w *world.World) {
	if r.Locked() {
		// Ignore this update; the repeater is locked.
		return
	}
	if r.inputStrength(pos, w) > 0 != r.Powered {
		w.ScheduleBlockUpdate(pos, time.Duration(r.Delay+1)*time.Millisecond*100)
	}
}

// ScheduledTick ...
func (r RedstoneRepeater) ScheduledTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	if r.Locked() {
		// Ignore this tick; the repeater is locked.
		return
	}

	if !r.Powered && r.inputStrength(pos, w) <= 0 {
		w.ScheduleBlockUpdate(pos, time.Duration(r.Delay+1)*time.Millisecond*100)
	}

	r.Powered = !r.Powered
	w.SetBlock(pos, r, &world.SetOpts{DisableBlockUpdates: true})
	updateSurroundingRedstone(pos, w)
}

// Locked ...
func (RedstoneRepeater) Locked() bool {
	//TODO implement me
	return false
}

// Source ...
func (r RedstoneRepeater) Source() bool {
	return r.Powered
}

// WeakPower ...
func (r RedstoneRepeater) WeakPower(_ cube.Pos, face cube.Face, _ *world.World, _ bool) int {
	if r.Powered && face == r.Facing.Face() {
		return 15
	}
	return 0
}

// StrongPower ...
func (r RedstoneRepeater) StrongPower(pos cube.Pos, face cube.Face, w *world.World, includeDust bool) int {
	return r.WeakPower(pos, face, w, includeDust)
}

// inputStrength ...
func (r RedstoneRepeater) inputStrength(pos cube.Pos, w *world.World) int {
	sidePos := pos.Side(r.Facing.Face())
	if strength := w.EmittedRedstonePower(sidePos, r.Facing.Face(), true); strength > 0 {
		return strength
	}
	if w, ok := w.Block(sidePos).(RedstoneWire); ok {
		return w.Power
	}
	return 0
}

// allRedstoneRepeaters ...
func allRedstoneRepeaters() (repeaters []world.Block) {
	for _, d := range cube.Directions() {
		for _, p := range []bool{false, true} {
			for i := 0; i < 4; i++ {
				repeaters = append(repeaters, RedstoneRepeater{Facing: d, Delay: i, Powered: p})
			}
		}
	}
	return
}

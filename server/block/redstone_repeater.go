package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
	"time"
)

// RedstoneRepeater is a block used in redstone circuits to "repeat" redstone signals back to full strength, delay
// signals, prevent signals moving backwards, or to "lock" signals in one state.
type RedstoneRepeater struct {
	transparent
	flowingWaterDisplacer

	// Facing is the direction from the torch to the block.
	Facing cube.Direction
	// Powered is true if the repeater is powered by a redstone signal.
	Powered bool
	// Delay represents the delay of the repeater in redstone ticks. It is between the range of one to four.
	Delay int
}

// SideClosed ...
func (RedstoneRepeater) SideClosed(cube.Pos, cube.Pos, *world.World) bool {
	return false
}

// Model ...
func (RedstoneRepeater) Model() world.BlockModel {
	return model.Diode{}
}

// BreakInfo ...
func (r RedstoneRepeater) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(r)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		updateGateRedstone(pos, tx, r.Facing.Face())
	})
}

// EncodeItem ...
func (RedstoneRepeater) EncodeItem() (name string, meta int16) {
	return "minecraft:repeater", 0
}

// EncodeBlock ...
func (r RedstoneRepeater) EncodeBlock() (string, map[string]any) {
	name := "minecraft:unpowered_repeater"
	if r.Powered {
		name = "minecraft:powered_repeater"
	}
	return name, map[string]any{
		"minecraft:cardinal_direction": r.Facing.String(),
		"repeater_delay":               int32(r.Delay),
	}
}

// UseOnBlock ...
func (r RedstoneRepeater) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, r)
	if !used {
		return false
	}
	b := tx.Block(pos.Side(cube.FaceDown))
	if d, ok := b.(LightDiffuser); ok && d.LightDiffusionLevel() == 0 {
		if _, isPiston := b.(Piston); !isPiston {
			return false
		}
	}

	r.Facing = user.Rotation().Direction().Opposite()

	place(tx, pos, r, user, ctx)
	if placed(ctx) {
		r.RedstoneUpdate(pos, tx)
		return true
	}
	return false
}

// NeighbourUpdateTick ...
func (r RedstoneRepeater) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	b := tx.Block(pos.Side(cube.FaceDown))
	if d, ok := b.(LightDiffuser); ok && d.LightDiffusionLevel() == 0 {
		if _, piston := b.(Piston); !piston {
			breakBlock(r, pos, tx)
		}
	}
}

// Activate ...
func (r RedstoneRepeater) Activate(pos cube.Pos, _ cube.Face, tx *world.Tx, _ item.User, _ *item.UseContext) bool {
	if r.Delay++; r.Delay > 3 {
		r.Delay = 0
	}
	tx.SetBlock(pos, r, nil)
	return true
}

// RedstoneUpdate ...
func (r RedstoneRepeater) RedstoneUpdate(pos cube.Pos, tx *world.Tx) {
	if r.Locked(pos, tx) {
		// Ignore this update; the repeater is locked.
		return
	}
	if r.inputStrength(pos, tx) > 0 != r.Powered {
		tx.ScheduleBlockUpdate(pos, r, time.Duration(r.Delay+1)*time.Millisecond*100)
	}
}

// ScheduledTick ...
func (r RedstoneRepeater) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if r.Locked(pos, tx) {
		// Ignore this tick; the repeater is locked.
		return
	}

	r.Powered = !r.Powered
	tx.SetBlock(pos, r, nil)
	updateGateRedstone(pos, tx, r.Facing.Face().Opposite())

	if r.Powered && r.inputStrength(pos, tx) <= 0 {
		tx.ScheduleBlockUpdate(pos, r, time.Duration(r.Delay+1)*time.Millisecond*100)
	}
	tx.SetBlock(pos, r, nil)
	updateGateRedstone(pos, tx, r.Facing.Face())
}

// Locked checks if the repeater is locked.
func (r RedstoneRepeater) Locked(pos cube.Pos, tx *world.Tx) bool {
	return r.locking(pos.Side(r.Facing.RotateLeft().Face()), r.Facing.RotateLeft(), tx) || r.locking(pos.Side(r.Facing.RotateRight().Face()), r.Facing.RotateRight(), tx)
}

// locking checks of the block at the given position is a powered repeater or comparator facing the repeater
func (r RedstoneRepeater) locking(pos cube.Pos, direction cube.Direction, tx *world.Tx) bool {
	block := tx.Block(pos)

	if repeater, ok := block.(RedstoneRepeater); ok {
		return repeater.Powered && repeater.Facing == direction
	}

	if comparator, ok := block.(RedstoneComparator); ok {
		return comparator.Powered && comparator.Facing == direction
	}
	return false
}

// RedstoneSource ...
func (r RedstoneRepeater) RedstoneSource() bool {
	return r.Powered
}

// WeakPower ...
func (r RedstoneRepeater) WeakPower(_ cube.Pos, face cube.Face, _ *world.Tx, _ bool) int {
	if r.Powered && face == r.Facing.Face() {
		return 15
	}
	return 0
}

// StrongPower ...
func (r RedstoneRepeater) StrongPower(pos cube.Pos, face cube.Face, tx *world.Tx, accountForDust bool) int {
	return r.WeakPower(pos, face, tx, accountForDust)
}

// inputStrength ...
func (r RedstoneRepeater) inputStrength(pos cube.Pos, tx *world.Tx) int {
	face := r.Facing.Face()
	return tx.RedstonePower(pos.Side(face), face, true)
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

package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand/v2"
	"time"
)

// Observer is a block that emits a redstone signal when the block or fluid it faces experiences a change.
type Observer struct {
	solid

	// Facing is the direction the observer is observing.
	Facing cube.Face
	// Powered is whether the observer is powered or not.
	Powered bool
}

// RedstoneSource ...
func (Observer) RedstoneSource() bool {
	return true
}

// RedstoneBlocking ...
func (Observer) RedstoneBlocking() bool {
	return true
}

// WeakPower ...
func (o Observer) WeakPower(pos cube.Pos, face cube.Face, tx *world.Tx, accountForDust bool) int {
	return o.StrongPower(pos, face, tx, accountForDust)
}

// StrongPower ...
func (o Observer) StrongPower(_ cube.Pos, face cube.Face, _ *world.Tx, _ bool) int {
	if !o.Powered || face != o.Facing {
		return 0
	}
	return 15
}

// ScheduledTick ...
func (o Observer) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	o.Powered = !o.Powered
	if o.Powered {
		tx.ScheduleBlockUpdate(pos, o, time.Millisecond*100)
	}
	tx.SetBlock(pos, o, nil)
	updateDirectionalRedstone(pos, tx, o.Facing.Opposite())
}

// NeighbourUpdateTick ...
func (o Observer) NeighbourUpdateTick(pos, changedNeighbour cube.Pos, tx *world.Tx) {
	if pos.Side(o.Facing) != changedNeighbour {
		return
	}
	if o.Powered {
		return
	}
	tx.ScheduleBlockUpdate(pos, o, time.Millisecond*100)
}

// UseOnBlock ...
func (o Observer) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, o)
	if !used {
		return false
	}
	o.Facing = calculateFace(user, pos, false)
	if o.Facing.Axis() == cube.Y {
		o.Facing = o.Facing.Opposite()
	}

	place(tx, pos, o, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (o Observer) BreakInfo() BreakInfo {
	return newBreakInfo(3, pickaxeHarvestable, pickaxeEffective, oneOf(o)).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		updateDirectionalRedstone(pos, tx, o.Facing.Opposite())
	})
}

// EncodeItem ...
func (o Observer) EncodeItem() (name string, meta int16) {
	return "minecraft:observer", 0
}

// EncodeBlock ...
func (o Observer) EncodeBlock() (string, map[string]any) {
	return "minecraft:observer", map[string]any{"minecraft:facing_direction": o.Facing.String(), "powered_bit": o.Powered}
}

// allObservers ...
func allObservers() (observers []world.Block) {
	for _, f := range cube.Faces() {
		observers = append(observers, Observer{Facing: f})
		observers = append(observers, Observer{Facing: f, Powered: true})
	}
	return
}

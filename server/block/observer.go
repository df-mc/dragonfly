package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"math/rand"
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

// Source ...
func (d Observer) Source() bool {
	return true
}

// WeakPower ...
func (d Observer) WeakPower(pos cube.Pos, face cube.Face, w *world.World, accountForDust bool) int {
	return d.StrongPower(pos, face, w, accountForDust)
}

// StrongPower ...
func (d Observer) StrongPower(_ cube.Pos, face cube.Face, w *world.World, _ bool) int {
	if !d.Powered || face != d.Facing {
		return 0
	}
	return 15
}

// ScheduledTick ...
func (d Observer) ScheduledTick(pos cube.Pos, w *world.World, _ *rand.Rand) {
	if !d.Powered {
		w.ScheduleBlockUpdate(pos, time.Millisecond*100)
	}
	d.Powered = !d.Powered
	w.SetBlock(pos, d, nil)
	updateDirectionalRedstone(pos, w, d.Facing.Opposite())
}

// NeighbourUpdateTick ...
func (d Observer) NeighbourUpdateTick(pos, changedNeighbour cube.Pos, w *world.World) {
	if pos.Side(d.Facing) != changedNeighbour {
		return
	}
	if d.Powered {
		return
	}
	w.ScheduleBlockUpdate(pos, time.Millisecond*100)
	updateDirectionalRedstone(pos, w, d.Facing.Opposite())
}

// UseOnBlock ...
func (d Observer) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(w, pos, face, d)
	if !used {
		return false
	}
	d.Facing = calculateAnySidedFace(user, pos, false)

	place(w, pos, d, user, ctx)
	return placed(ctx)
}

// BreakInfo ...
func (d Observer) BreakInfo() BreakInfo {
	return newBreakInfo(3, pickaxeHarvestable, pickaxeEffective, oneOf(d)).withBreakHandler(func(pos cube.Pos, w *world.World, _ item.User) {
		updateDirectionalRedstone(pos, w, d.Facing.Opposite())
	})
}

// EncodeItem ...
func (d Observer) EncodeItem() (name string, meta int16) {
	return "minecraft:observer", 0
}

// EncodeBlock ...
func (d Observer) EncodeBlock() (string, map[string]any) {
	return "minecraft:observer", map[string]any{"facing_direction": int32(d.Facing), "powered_bit": d.Powered}
}

// allObservers ...
func allObservers() (observers []world.Block) {
	for _, f := range cube.Faces() {
		observers = append(observers, Observer{Facing: f})
		observers = append(observers, Observer{Facing: f, Powered: true})
	}
	return
}

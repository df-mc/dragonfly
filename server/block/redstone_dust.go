package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"sync"
)

type RedstoneDust struct {
	empty
	transparent
	Power int
}

// EncodeItem ...
func (r RedstoneDust) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone", 0
}

// EncodeBlock ...
func (r RedstoneDust) EncodeBlock() (string, map[string]any) {
	return "minecraft:redstone_wire", map[string]any{
		"redstone_signal": int32(r.Power),
	}
}

// UseOnBlock ...
func (r RedstoneDust) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, r)
	if !used {
		return
	}
	belowPos := pos.Side(cube.FaceDown)
	if !w.Block(belowPos).Model().FaceSolid(belowPos, cube.FaceUp, w) {
		return
	}
	r.Power = r.receivedPower(pos, w)
	place(w, pos, r, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (r RedstoneDust) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	power := r.receivedPower(pos, w)
	if r.Power != power {
		r.Power = power
		w.SetBlock(pos, r, nil)
	}
}

// disabledEmitters ...
var disabledEmitters sync.Map

// WeakPower ...
func (r RedstoneDust) WeakPower(pos cube.Pos, face cube.Face, w *world.World) int {
	if _, ok := disabledEmitters.Load(pos.Side(face.Opposite())); ok {
		return 0
	}
	if face == cube.FaceDown {
		return 0
	}
	if face == cube.FaceUp {
		return r.Power
	}
	if r.connection(pos, face, w) && !r.connection(pos, face.RotateLeft(), w) && !r.connection(pos, face.RotateRight(), w) {
		return r.Power
	}
	return 0
}

// StrongPower ...
func (r RedstoneDust) StrongPower(pos cube.Pos, face cube.Face, w *world.World) int {
	return r.WeakPower(pos, face, w)
}

// receivedPower returns the highest level of received redstone power at the provided position.
func (r RedstoneDust) receivedPower(pos cube.Pos, w *world.World) int {
	disabledEmitters.Store(pos, struct{}{})
	received := w.ReceivedRedstonePower(pos)
	disabledEmitters.Delete(pos)

	var power int
	if received < 15 {
		_, solidAbove := w.Block(pos.Side(cube.FaceUp)).Model().(model.Solid)
		for _, face := range cube.HorizontalFaces() {
			sidePos := pos.Side(face)
			power = max(power, r.checkPower(sidePos, w))
			if _, sideSolid := w.Block(sidePos).Model().(model.Solid); sideSolid && !solidAbove {
				power = max(power, r.checkPower(sidePos.Side(cube.FaceUp), w))
			} else if !sideSolid {
				power = max(power, r.checkPower(sidePos.Side(cube.FaceDown), w))
			}
		}
	}
	return max(received, power-1)
}

// checkPower attempts to return the power level of the redstone dust at the provided position if it exists. If there is
// no redstone dust at the position, 0 is returned.
func (r RedstoneDust) checkPower(pos cube.Pos, w *world.World) int {
	if b, ok := w.Block(pos).(RedstoneDust); ok {
		return b.Power
	}
	return 0
}

// connection returns true if the dust connects to the given face.
func (r RedstoneDust) connection(pos cube.Pos, face cube.Face, w *world.World) bool {
	sidePos := pos.Side(face)
	sideBlock := w.Block(sidePos)
	if _, solidAbove := w.Block(pos.Side(cube.FaceUp)).Model().(model.Solid); !solidAbove && r.canRunOnTop(w, sidePos, sideBlock) && r.connectsTo(w.Block(sidePos.Side(cube.FaceUp)), false) {
		return true
	}
	_, sideSolid := sideBlock.Model().(model.Solid)
	return r.connectsTo(sideBlock, true) || !sideSolid && r.connectsTo(w.Block(sidePos.Side(cube.FaceDown)), false)
}

// connectsTo ...
func (r RedstoneDust) connectsTo(block world.Block, hasFace bool) bool {
	switch block.(type) {
	case RedstoneDust:
		return true
		// TODO: Repeaters, observers
	}
	if _, ok := block.(world.Conductor); ok {
		return hasFace
	}
	return false
}

// canRunOnTop ...
func (r RedstoneDust) canRunOnTop(w *world.World, pos cube.Pos, block world.Block) bool {
	// TODO: Hoppers.
	return block.Model().FaceSolid(pos, cube.FaceUp, w)
}

// allRedstoneDust returns a list of all redstone dust states.
func allRedstoneDust() (all []world.Block) {
	for i := 0; i < 16; i++ {
		all = append(all, RedstoneDust{Power: i})
	}
	return
}

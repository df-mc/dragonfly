package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type RedstoneDust struct {
	empty
	transparent
	Power    int
	emitting bool
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
	r.emitting = true
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

// WeakPower ...
func (r RedstoneDust) WeakPower(_ cube.Pos, face cube.Face, _ *world.World) int {
	if !r.emitting || face == cube.FaceDown {
		return 0
	}
	if r.Power > 0 {
		// TODO: Some connectivity logic
		return r.Power
	}
	return 0
}

// StrongPower ...
func (r RedstoneDust) StrongPower(pos cube.Pos, face cube.Face, w *world.World) int {
	if !r.emitting {
		return 0
	}
	return r.WeakPower(pos, face, w)
}

// receivedPower returns the highest level of received redstone power at the provided position.
func (r RedstoneDust) receivedPower(pos cube.Pos, w *world.World) int {
	r.emitting = false
	received := w.ReceivedRedstonePower(pos)
	r.emitting = true
	var power int
	if received < 15 {
		_, solidAbove := w.Block(pos.Side(cube.FaceUp)).Model().(model.Solid)
		for _, face := range cube.HorizontalFaces() {
			sidePos := pos.Side(face)
			received = max(received, r.checkPower(sidePos, w))
			_, sideSolid := w.Block(sidePos).Model().(model.Solid)
			if sideSolid && !solidAbove {
				received = max(received, r.checkPower(sidePos.Side(cube.FaceUp), w))
			} else if !sideSolid {
				received = max(received, r.checkPower(sidePos.Side(cube.FaceDown), w))
			}
		}
	}
	return max(power, received-1)
}

// checkPower attempts to return the power level of the redstone dust at the provided position if it exists. If there is
// no redstone dust at the position, 0 is returned.
func (r RedstoneDust) checkPower(pos cube.Pos, w *world.World) int {
	block := w.Block(pos)
	if b, ok := block.(RedstoneDust); ok {
		return b.Power
	}
	return 0
}

// allRedstoneDust returns a list of all redstone dust states.
func allRedstoneDust() (all []world.Block) {
	for i := 0; i < 16; i++ {
		all = append(all, RedstoneDust{Power: i})
	}
	return
}

package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/internal/sliceutil"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"golang.org/x/exp/slices"
)

type RedstoneDust struct {
	empty
	transparent
	Power           int
	disableEmitting bool
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
		for _, face := range cube.Faces() {
			sidePos := pos.Side(face)
			if n, ok := w.Block(sidePos).(world.Conductor); ok {
				n.NeighbourUpdateTick(sidePos, pos, w)
			}
		}
	}
}

// WeakPower ...
func (r RedstoneDust) WeakPower(pos cube.Pos, side cube.Face, w *world.World) int {
	if r.disableEmitting {
		return 0
	}
	if side.Axis() == cube.Y {
		return r.Power
	}

	faces := sliceutil.Filter(cube.HorizontalFaces(), func(face cube.Face) bool {
		return r.powers(pos, face, w)
	})
	if side.Axis() != cube.Y && len(faces) == 0 {
		return r.Power
	} else if slices.Contains(faces, side) && !slices.Contains(faces, side.RotateLeft()) && !slices.Contains(faces, side.RotateRight()) {
		return r.Power
	}
	return 0
}

// StrongPower ...
func (r RedstoneDust) StrongPower(pos cube.Pos, face cube.Face, w *world.World) int {
	if r.disableEmitting {
		return 0
	}
	return r.WeakPower(pos, face, w)
}

// receivedPower returns the highest level of received redstone power at the provided position.
func (r RedstoneDust) receivedPower(pos cube.Pos, w *world.World) int {
	r.disableEmitting = true
	received := w.ReceivedRedstonePower(pos)
	r.disableEmitting = false

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
	block := w.Block(pos)
	if b, ok := block.(RedstoneDust); ok {
		return b.Power
	}
	return 0
}

// powers returns true if the dust powers the given face.
func (r RedstoneDust) powers(pos cube.Pos, face cube.Face, w *world.World) bool {
	// TODO
	return true
}

// allRedstoneDust returns a list of all redstone dust states.
func allRedstoneDust() (all []world.Block) {
	for i := 0; i < 16; i++ {
		all = append(all, RedstoneDust{Power: i})
	}
	return
}

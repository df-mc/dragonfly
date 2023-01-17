package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

type RedstoneWire struct {
	empty
	transparent
	Power int
}

// EncodeItem ...
func (r RedstoneWire) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone", 0
}

// EncodeBlock ...
func (r RedstoneWire) EncodeBlock() (string, map[string]any) {
	return "minecraft:redstone_wire", map[string]any{
		"redstone_signal": int32(r.Power),
	}
}

// UseOnBlock ...
func (r RedstoneWire) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, w *world.World, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(w, pos, face, r)
	if !used {
		return
	}
	belowPos := pos.Side(cube.FaceDown)
	if !w.Block(belowPos).Model().FaceSolid(belowPos, cube.FaceUp, w) {
		return
	}
	r.Power = r.calculatePower(pos, w)
	place(w, pos, r, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (r RedstoneWire) NeighbourUpdateTick(pos, _ cube.Pos, w *world.World) {
	power := r.calculatePower(pos, w)
	if r.Power != power {
		r.Power = power
		w.SetBlock(pos, r, nil)
		updateSurroundingRedstone(pos, w)
	}
}

// WeakPower ...
func (r RedstoneWire) WeakPower(pos cube.Pos, face cube.Face, w *world.World, includeDust bool) int {
	if !includeDust {
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
func (r RedstoneWire) StrongPower(pos cube.Pos, face cube.Face, w *world.World, includeDust bool) int {
	return r.WeakPower(pos, face, w, includeDust)
}

// calculatePower returns the highest level of received redstone power at the provided position.
func (r RedstoneWire) calculatePower(pos cube.Pos, w *world.World) int {
	aboveBlock := w.Block(pos.Side(cube.FaceUp))
	_, aboveSolid := aboveBlock.Model().(model.Solid)

	var blockPower, wirePower int
	for _, side := range cube.Faces() {
		neighbourPos := pos.Side(side)
		neighbour := w.Block(neighbourPos)

		wirePower = r.maxCurrentStrength(wirePower, neighbourPos, w)
		blockPower = max(blockPower, w.EmittedRedstonePower(neighbourPos, side, false))

		if side.Axis() == cube.Y {
			// Only check horizontal neighbours from here on.
			continue
		}

		if d, ok := neighbour.(LightDiffuser); (!ok || d.LightDiffusionLevel() > 0) && !aboveSolid {
			wirePower = r.maxCurrentStrength(wirePower, neighbourPos.Side(cube.FaceUp), w)
		}
		if _, neighbourSolid := neighbour.Model().(model.Solid); !neighbourSolid {
			wirePower = r.maxCurrentStrength(wirePower, neighbourPos.Side(cube.FaceDown), w)
		}
	}
	return max(blockPower, wirePower-1)
}

// maxCurrentStrength ...
func (r RedstoneWire) maxCurrentStrength(power int, pos cube.Pos, w *world.World) int {
	if wire, ok := w.Block(pos).(RedstoneWire); ok {
		return max(wire.Power, power)
	}
	return power
}

// connection returns true if the dust connects to the given face.
func (r RedstoneWire) connection(pos cube.Pos, face cube.Face, w *world.World) bool {
	sidePos := pos.Side(face)
	sideBlock := w.Block(sidePos)
	if _, solidAbove := w.Block(pos.Side(cube.FaceUp)).Model().(model.Solid); !solidAbove && r.canRunOnTop(w, sidePos, sideBlock) && r.connectsTo(w.Block(sidePos.Side(cube.FaceUp)), false) {
		return true
	}
	_, sideSolid := sideBlock.Model().(model.Solid)
	return r.connectsTo(sideBlock, true) || !sideSolid && r.connectsTo(w.Block(sidePos.Side(cube.FaceDown)), false)
}

// connectsTo ...
func (r RedstoneWire) connectsTo(block world.Block, hasFace bool) bool {
	switch block.(type) {
	case RedstoneWire:
		return true
		// TODO: Repeaters, observers
	}
	if _, ok := block.(world.Conductor); ok {
		return hasFace
	}
	return false
}

// canRunOnTop ...
func (r RedstoneWire) canRunOnTop(w *world.World, pos cube.Pos, block world.Block) bool {
	// TODO: Hoppers.
	return block.Model().FaceSolid(pos, cube.FaceUp, w)
}

// allRedstoneDust returns a list of all redstone dust states.
func allRedstoneDust() (all []world.Block) {
	for i := 0; i < 16; i++ {
		all = append(all, RedstoneWire{Power: i})
	}
	return
}

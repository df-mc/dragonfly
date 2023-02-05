package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/block/model"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// RedstoneWire is a block that is used to transfer a charge between objects. Charged objects can be used to open doors
// or activate certain items. This block is the placed form of redstone which can be found by mining Redstone Ore with
// an Iron Pickaxe or better. Deactivated redstone wire will appear dark red, but activated redstone wire will appear
// bright red with a sparkling particle effect.
type RedstoneWire struct {
	empty
	transparent

	// Power is the current power level of the redstone wire. It ranges from 0 to 15.
	Power int
}

// HasLiquidDrops ...
func (RedstoneWire) HasLiquidDrops() bool {
	return true
}

// BreakInfo ...
func (r RedstoneWire) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(r)).withBreakHandler(func(pos cube.Pos, w *world.World, _ item.User) {
		updateStrongRedstone(pos, w)
	})
}

// EncodeItem ...
func (RedstoneWire) EncodeItem() (name string, meta int16) {
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
	if _, ok := w.Liquid(pos); ok {
		return false
	}
	belowPos := pos.Side(cube.FaceDown)
	if !w.Block(belowPos).Model().FaceSolid(belowPos, cube.FaceUp, w) {
		return
	}
	r.Power = r.calculatePower(pos, w)
	place(w, pos, r, user, ctx)
	if placed(ctx) {
		updateStrongRedstone(pos, w)
		return true
	}
	return false
}

// NeighbourUpdateTick ...
func (r RedstoneWire) NeighbourUpdateTick(pos, neighbour cube.Pos, w *world.World) {
	if pos == neighbour {
		// Ignore neighbour updates on ourself.
		return
	}
	if _, ok := w.Block(pos.Side(cube.FaceDown)).(Air); ok {
		w.SetBlock(pos, nil, nil)
		dropItem(w, item.NewStack(r, 1), pos.Vec3Centre())
		return
	}
	r.RedstoneUpdate(pos, w)
}

// RedstoneUpdate ...
func (r RedstoneWire) RedstoneUpdate(pos cube.Pos, w *world.World) {
	if power := r.calculatePower(pos, w); r.Power != power {
		r.Power = power
		w.SetBlock(pos, r, &world.SetOpts{DisableBlockUpdates: true})
		updateStrongRedstone(pos, w)
	}
}

// Source ...
func (RedstoneWire) Source() bool {
	return false
}

// WeakPower ...
func (r RedstoneWire) WeakPower(pos cube.Pos, face cube.Face, w *world.World, accountForDust bool) int {
	if !accountForDust {
		return 0
	}
	if face == cube.FaceUp {
		return r.Power
	}
	if face == cube.FaceDown {
		return 0
	}
	if r.connection(pos, face.Opposite(), w) {
		return r.Power
	}
	if r.connection(pos, face, w) && !r.connection(pos, face.RotateLeft(), w) && !r.connection(pos, face.RotateRight(), w) {
		return r.Power
	}
	return 0
}

// StrongPower ...
func (r RedstoneWire) StrongPower(pos cube.Pos, face cube.Face, w *world.World, accountForDust bool) int {
	return r.WeakPower(pos, face, w, accountForDust)
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
		blockPower = max(blockPower, w.RedstonePower(neighbourPos, side, false))

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
func (RedstoneWire) maxCurrentStrength(power int, pos cube.Pos, w *world.World) int {
	if wire, ok := w.Block(pos).(RedstoneWire); ok {
		return max(wire.Power, power)
	}
	return power
}

// connection returns true if the dust connects to the given face.
func (r RedstoneWire) connection(pos cube.Pos, face cube.Face, w *world.World) bool {
	sidePos := pos.Side(face)
	sideBlock := w.Block(sidePos)
	if _, solidAbove := w.Block(pos.Side(cube.FaceUp)).Model().(model.Solid); !solidAbove && r.canRunOnTop(w, sidePos, sideBlock) && r.connectsTo(w.Block(sidePos.Side(cube.FaceUp)), face, false) {
		return true
	}
	_, sideSolid := sideBlock.Model().(model.Solid)
	return r.connectsTo(sideBlock, face, true) || !sideSolid && r.connectsTo(w.Block(sidePos.Side(cube.FaceDown)), face, false)
}

// connectsTo ...
func (RedstoneWire) connectsTo(block world.Block, face cube.Face, allowDirectSources bool) bool {
	switch r := block.(type) {
	case RedstoneWire:
		return true
	case RedstoneRepeater:
		return r.Facing.Face() == face || r.Facing.Face().Opposite() == face
	case Piston:
		return true
	}
	// TODO: Account for observers.
	c, ok := block.(world.Conductor)
	return ok && allowDirectSources && c.Source()
}

// canRunOnTop ...
func (RedstoneWire) canRunOnTop(w *world.World, pos cube.Pos, block world.Block) bool {
	// TODO: Hoppers.
	return block.Model().FaceSolid(pos, cube.FaceUp, w)
}

// allRedstoneWires returns a list of all redstone dust states.
func allRedstoneWires() (all []world.Block) {
	for i := 0; i < 16; i++ {
		all = append(all, RedstoneWire{Power: i})
	}
	return
}

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
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(RedstoneWire{})).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		updateStrongRedstone(pos, tx)
	})
}

// EncodeBlock ...
func (r RedstoneWire) EncodeBlock() (string, map[string]any) {
	return "minecraft:redstone_wire", map[string]any{
		"redstone_signal": int32(r.Power),
	}
}

// EncodeItem ...
func (RedstoneWire) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone", 0
}

// UseOnBlock ...
func (r RedstoneWire) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) (used bool) {
	pos, _, used = firstReplaceable(tx, pos, face, r)
	if !used {
		return
	}
	belowPos := pos.Side(cube.FaceDown)
	if !tx.Block(belowPos).Model().FaceSolid(belowPos, cube.FaceUp, tx) {
		return
	}
	r.Power = r.calculatePower(pos, tx)
	place(tx, pos, r, user, ctx)
	if placed(ctx) {
		updateStrongRedstone(pos, tx)
		return true
	}
	return false
}

// NeighbourUpdateTick ...
func (r RedstoneWire) NeighbourUpdateTick(pos, neighbour cube.Pos, tx *world.Tx) {
	if pos == neighbour {
		// Ignore the self-update sent after this wire's block state changes.
		return
	}
	below := pos.Side(cube.FaceDown)
	if !tx.Block(below).Model().FaceSolid(below, cube.FaceUp, tx) {
		breakBlock(r, pos, tx)
		return
	}
	updateRedstone(pos, tx)
}

// RedstoneUpdate ...
func (r RedstoneWire) RedstoneUpdate(pos cube.Pos, tx *world.Tx) {
	if power := r.calculatePower(pos, tx); r.Power != power {
		r.Power = power
		tx.SetBlock(pos, r, &world.SetOpts{DisableBlockUpdates: true})
		updateStrongRedstone(pos, tx)
	}
}

// RedstoneSource ...
func (RedstoneWire) RedstoneSource() bool {
	return false
}

// WeakPower returns the power emitted by the wire toward a neighbouring receiver. Dust powers upward, never powers
// downward, and only powers horizontal receivers in connected directions. A powered wire with no horizontal
// connections behaves as an unconnected cross and powers every horizontal side.
func (r RedstoneWire) WeakPower(pos cube.Pos, face cube.Face, tx *world.Tx, accountForDust bool) int {
	if !accountForDust {
		return 0
	}
	if face == cube.FaceUp {
		return r.Power
	}
	if face == cube.FaceDown {
		return 0
	}
	if !r.hasHorizontalRedstoneConnection(pos, tx) {
		return r.Power
	}
	if r.connection(pos, face.Opposite(), tx) {
		return r.Power
	}
	if r.connection(pos, face, tx) && !r.connection(pos, face.RotateLeft(), tx) && !r.connection(pos, face.RotateRight(), tx) {
		return r.Power
	}
	return 0
}

// StrongPower ...
func (r RedstoneWire) StrongPower(pos cube.Pos, face cube.Face, tx *world.Tx, accountForDust bool) int {
	return r.WeakPower(pos, face, tx, accountForDust)
}

// calculatePower returns the highest level of received redstone power at the provided position.
func (r RedstoneWire) calculatePower(pos cube.Pos, tx *world.Tx) int {
	aboveBlock := tx.Block(pos.Side(cube.FaceUp))
	aboveBlocksVerticalTravel := blocksRedstoneWireVerticalTravel(aboveBlock)

	var blockPower, wirePower int
	for _, side := range cube.Faces() {
		neighbourPos := pos.Side(side)
		neighbour := tx.Block(neighbourPos)

		wirePower = r.maxCurrentStrength(wirePower, neighbourPos, tx)
		blockPower = max(blockPower, tx.RedstonePower(neighbourPos, side, false))

		if side.Axis() == cube.Y {
			// Only check horizontal neighbors from here on.
			continue
		}

		if canRedstoneWireStepDown(pos, neighbourPos, neighbour, tx) && !aboveBlocksVerticalTravel {
			wirePower = r.maxCurrentStrength(wirePower, neighbourPos.Side(cube.FaceUp), tx)
		}
		if canRedstoneWireStepDown(neighbourPos.Side(cube.FaceDown), neighbourPos, neighbour, tx) && !blocksRedstoneWireVerticalTravel(neighbour) {
			wirePower = r.maxCurrentStrength(wirePower, neighbourPos.Side(cube.FaceDown), tx)
		}

		if _, neighbourSolid := neighbour.Model().(model.Solid); !neighbourSolid {
			wirePower = r.maxCurrentStrength(wirePower, neighbourPos.Side(cube.FaceDown), tx)
		}
	}
	return max(blockPower, wirePower-1)
}

// maxCurrentStrength ...
func (RedstoneWire) maxCurrentStrength(power int, pos cube.Pos, tx *world.Tx) int {
	return maxRedstoneWirePower(tx.Block(pos), power)
}

// hasHorizontalRedstoneConnection checks if the dust connects horizontally to redstone wire or a redstone source. It
// does not include passive receivers such as doors, trapdoors, or note blocks.
func (r RedstoneWire) hasHorizontalRedstoneConnection(pos cube.Pos, tx *world.Tx) bool {
	for _, face := range cube.HorizontalFaces() {
		if r.connection(pos, face, tx) {
			return true
		}
	}
	return false
}

// connection returns true if the dust shape connects through the given face to another wire or a redstone source. It
// also accounts for valid one-block vertical wire connections.
func (r RedstoneWire) connection(pos cube.Pos, face cube.Face, tx *world.Tx) bool {
	sidePos := pos.Side(face)
	sideBlock := tx.Block(sidePos)
	if !blocksRedstoneWireVerticalTravel(tx.Block(pos.Side(cube.FaceUp))) && r.canRunOnTop(tx, sidePos, sideBlock) && r.connectsTo(tx.Block(sidePos.Side(cube.FaceUp)), false) {
		return true
	}
	_, sideSolid := sideBlock.Model().(model.Solid)
	return r.connectsTo(sideBlock, true) || !sideSolid && r.connectsTo(tx.Block(sidePos.Side(cube.FaceDown)), false)
}

// connectsTo reports whether a block is part of the redstone wire connection graph. Passive redstone receivers are not
// connections; direct source conductors count only when allowDirectSources is true.
func (RedstoneWire) connectsTo(block world.Block, allowDirectSources bool) bool {
	if _, ok := block.(RedstoneWire); ok {
		return true
	}
	c, ok := block.(world.Conductor)
	return ok && allowDirectSources && c.RedstoneSource()
}

// canRunOnTop checks whether redstone dust can be placed on top of the block.
func (RedstoneWire) canRunOnTop(tx *world.Tx, pos cube.Pos, block world.Block) bool {
	return block.Model().FaceSolid(pos, cube.FaceUp, tx)
}

// blocksRedstoneWireVerticalTravel checks if the block above redstone wire blocks vertical wire travel.
func blocksRedstoneWireVerticalTravel(block world.Block) bool {
	if _, ok := block.Model().(model.Solid); !ok {
		return false
	}
	diffuser, ok := block.(LightDiffuser)
	return !ok || diffuser.LightDiffusionLevel() != 0
}

// canRedstoneWireStepDown checks if redstone dust can provide power while travelling down around the side block.
func canRedstoneWireStepDown(from, side cube.Pos, block world.Block, tx *world.Tx) bool {
	if stepDowner, ok := block.(RedstoneWireStepDowner); ok {
		return stepDowner.CanRedstoneWireStepDown(side, from, tx)
	}
	for _, face := range cube.Faces() {
		if !block.Model().FaceSolid(side, face, tx) {
			return false
		}
	}
	return true
}

// TrimMaterial ...
func (RedstoneWire) TrimMaterial() string {
	return item.RedstoneWire{}.TrimMaterial()
}

// MaterialColour ...
func (RedstoneWire) MaterialColour() string {
	return item.RedstoneWire{}.MaterialColour()
}

// allRedstoneWires returns a list of all redstone dust states.
func allRedstoneWires() (all []world.Block) {
	for i := 0; i < 16; i++ {
		all = append(all, RedstoneWire{Power: i})
	}
	return
}

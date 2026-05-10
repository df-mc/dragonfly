package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// RedstoneBlock is a solid block that emits maximum redstone power.
type RedstoneBlock struct {
	solid
}

// BreakInfo ...
func (r RedstoneBlock) BreakInfo() BreakInfo {
	return newBreakInfo(5, pickaxeHarvestable, pickaxeEffective, oneOf(r)).withBlastResistance(30)
}

// RedstonePower always returns maximum power.
func (RedstoneBlock) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	return 15
}

// RedstoneStrongPower returns no strong power. Redstone blocks power adjacent components directly, but do not power
// adjacent opaque blocks.
func (RedstoneBlock) RedstoneStrongPower(cube.Pos, *world.Tx, cube.Face) int {
	return 0
}

// EncodeItem ...
func (RedstoneBlock) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_block", 0
}

// EncodeBlock ...
func (RedstoneBlock) EncodeBlock() (string, map[string]any) {
	return "minecraft:redstone_block", nil
}

// RedstoneWire is redstone dust placed in the world. Power is stored as a value from 0 to 15.
type RedstoneWire struct {
	empty
	transparent
	sourceWaterDisplacer

	// Power is the current signal strength carried by the wire.
	Power int
}

// UseOnBlock places redstone wire on a replaceable block.
func (r RedstoneWire) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, r)
	if !used || !redstoneWireSupported(tx, pos) {
		return false
	}
	place(tx, pos, r, user, ctx)
	return placed(ctx)
}

// RedstonePower returns the wire's current signal strength.
func (r RedstoneWire) RedstonePower(cube.Pos, *world.Tx, cube.Face) int {
	return r.Power
}

// RedstoneSignalLoss returns the signal loss through a wire segment.
func (RedstoneWire) RedstoneSignalLoss(cube.Pos, *world.Tx, cube.Face, cube.Face) int {
	return 1
}

// RedstoneRelayerNeighbours returns all wire positions directly connected to this dust, including dust stepping up or
// down adjacent blocks.
func (RedstoneWire) RedstoneRelayerNeighbours(pos cube.Pos, tx *world.Tx) []cube.Pos {
	neighbours := make([]cube.Pos, 0, 12)
	for _, face := range cube.HorizontalFaces() {
		side := pos.Side(face)
		if side.OutOfBounds(tx.Range()) {
			continue
		}
		neighbours = append(neighbours, side)

		above := pos.Side(cube.FaceUp)
		if !redstoneWireBlocksConnectionLoaded(tx, above, cube.FaceDown) && redstoneWireSupportedLoaded(tx, side.Side(cube.FaceUp)) {
			neighbours = append(neighbours, side.Side(cube.FaceUp))
		}
		if !redstoneWireBlocksConnectionLoaded(tx, side, cube.FaceUp) {
			down := side.Side(cube.FaceDown)
			if !down.OutOfBounds(tx.Range()) {
				neighbours = append(neighbours, down)
			}
		}
	}
	return neighbours
}

// RedstonePowerUpdate updates the wire strength to match its strongest input.
func (r RedstoneWire) RedstonePowerUpdate(_ cube.Pos, _ *world.Tx, power int) (world.Block, bool) {
	power = max(0, min(power, 15))
	if r.Power == power {
		return r, false
	}
	r.Power = power
	return r, true
}

// NeighbourUpdateTick breaks unsupported wire.
func (r RedstoneWire) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !redstoneWireSupported(tx, pos) {
		breakBlock(r, pos, tx)
	}
}

// HasLiquidDrops ...
func (RedstoneWire) HasLiquidDrops() bool {
	return true
}

// BreakInfo ...
func (r RedstoneWire) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(RedstoneWire{}))
}

// SideClosed ...
func (RedstoneWire) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

// EncodeItem ...
func (RedstoneWire) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone", 0
}

// EncodeBlock ...
func (r RedstoneWire) EncodeBlock() (string, map[string]any) {
	return "minecraft:redstone_wire", map[string]any{"redstone_signal": int32(max(0, min(r.Power, 15)))}
}

func allRedstoneWires() (wires []world.Block) {
	for i := 0; i <= 15; i++ {
		wires = append(wires, RedstoneWire{Power: i})
	}
	return
}

func redstoneWireSupported(tx *world.Tx, pos cube.Pos) bool {
	below := pos.Side(cube.FaceDown)
	if below.OutOfBounds(tx.Range()) {
		return false
	}
	return tx.Block(below).Model().FaceSolid(below, cube.FaceUp, tx)
}

func redstoneWireSupportedLoaded(tx *world.Tx, pos cube.Pos) bool {
	below := pos.Side(cube.FaceDown)
	if below.OutOfBounds(tx.Range()) {
		return false
	}
	b, ok := tx.BlockLoaded(below)
	return ok && b.Model().FaceSolid(below, cube.FaceUp, tx)
}

func redstoneWireBlocksConnectionLoaded(tx *world.Tx, pos cube.Pos, face cube.Face) bool {
	if pos.OutOfBounds(tx.Range()) {
		return true
	}
	b, ok := tx.BlockLoaded(pos)
	return ok && b.Model().FaceSolid(pos, face, tx)
}

// RedstoneLamp is a lamp that lights while powered.
type RedstoneLamp struct {
	solid

	// Lit is true when the lamp is powered and emitting light.
	Lit bool
}

// LightEmissionLevel ...
func (r RedstoneLamp) LightEmissionLevel() uint8 {
	if r.Lit {
		return 15
	}
	return 0
}

// RedstonePowerUpdate updates the lamp's lit state to match its redstone input.
func (r RedstoneLamp) RedstonePowerUpdate(_ cube.Pos, _ *world.Tx, power int) (world.Block, bool) {
	lit := power > 0
	if r.Lit == lit {
		return r, false
	}
	r.Lit = lit
	return r, true
}

// BreakInfo ...
func (r RedstoneLamp) BreakInfo() BreakInfo {
	return newBreakInfo(0.3, alwaysHarvestable, nothingEffective, oneOf(RedstoneLamp{}))
}

// EncodeItem ...
func (RedstoneLamp) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone_lamp", 0
}

// EncodeBlock ...
func (r RedstoneLamp) EncodeBlock() (string, map[string]any) {
	if r.Lit {
		return "minecraft:lit_redstone_lamp", nil
	}
	return "minecraft:redstone_lamp", nil
}

func allRedstoneLamps() (lamps []world.Block) {
	return []world.Block{RedstoneLamp{}, RedstoneLamp{Lit: true}}
}

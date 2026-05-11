package block

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// RedstoneWire is a block that is used to transfer a charge between objects. Charged objects can be used to open doors
// or activate certain items. This block is the placed form of redstone which can be found by mining redstone ore with
// an iron pickaxe or better. Deactivated redstone wire will appear dark red, but activated redstone wire will appear
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
		tx.ScheduleRedstoneUpdate(pos)
	})
}

// EncodeBlock ...
func (r RedstoneWire) EncodeBlock() (string, map[string]any) {
	return "minecraft:redstone_wire", map[string]any{
		"redstone_signal": int32(redstonePower(r.Power)),
	}
}

// EncodeItem ...
func (RedstoneWire) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone", 0
}

// UseOnBlock ...
func (r RedstoneWire) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, r)
	if !used || !redstoneWireSupported(tx, pos) {
		return false
	}
	r.Power = tx.RedstonePower(pos)
	place(tx, pos, r, user, ctx)
	return placed(ctx)
}

// NeighbourUpdateTick ...
func (r RedstoneWire) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !redstoneWireSupported(tx, pos) {
		breakBlock(r, pos, tx)
	}
}

// RedstonePower returns the wire's current signal strength from connected faces.
func (r RedstoneWire) RedstonePower(pos cube.Pos, tx *world.Tx, face cube.Face) int {
	if face == cube.FaceDown {
		return 0
	}
	if tx != nil && redstoneWireFaceHorizontal(face) && !redstoneWirePowersHorizontalFace(pos, tx, face) {
		return 0
	}
	return redstonePower(r.Power)
}

// RedstoneSignalLoss returns the signal loss through a wire segment.
func (RedstoneWire) RedstoneSignalLoss(cube.Pos, *world.Tx, cube.Face, cube.Face) int {
	return 1
}

// RedstoneRelayerNeighbours returns all wire positions directly connected to this dust, including dust stepping up or
// down adjacent blocks.
func (RedstoneWire) RedstoneRelayerNeighbours(pos cube.Pos, tx *world.Tx) []cube.Pos {
	neighbours := make([]cube.Pos, 0, 12)
	faces := redstoneWirePoweredHorizontalFaces(pos, tx)
	for _, face := range cube.HorizontalFaces() {
		if !faces[face] {
			continue
		}
		side := pos.Side(face)
		if side.OutOfBounds(tx.Range()) {
			continue
		}
		positions := redstoneWireHorizontalConnectionPositions(pos, tx, face)
		if len(positions) != 0 {
			neighbours = append(neighbours, positions...)
			continue
		}
		if redstoneWireRelevantLoaded(tx, side) {
			neighbours = append(neighbours, side)
		}
	}
	return neighbours
}

// RedstonePowerUpdate updates the wire strength to match its strongest input.
func (r RedstoneWire) RedstonePowerUpdate(_ cube.Pos, _ *world.Tx, power int) (world.Block, bool) {
	power = redstonePower(power)
	if r.Power == power {
		return r, false
	}
	r.Power = power
	return r, true
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

func redstoneWirePowersHorizontalFace(pos cube.Pos, tx *world.Tx, face cube.Face) bool {
	return redstoneWirePoweredHorizontalFaces(pos, tx)[face]
}

func redstoneWirePoweredHorizontalFaces(pos cube.Pos, tx *world.Tx) map[cube.Face]bool {
	connections := make(map[cube.Face]bool, len(cube.HorizontalFaces()))
	for _, face := range cube.HorizontalFaces() {
		if len(redstoneWireHorizontalConnectionPositions(pos, tx, face)) != 0 {
			connections[face] = true
		}
	}
	switch len(connections) {
	case 0:
		for _, face := range cube.HorizontalFaces() {
			connections[face] = true
		}
	case 1:
		for face := range connections {
			connections[face.Opposite()] = true
		}
	}
	return connections
}

func redstoneWireHorizontalConnectionPositions(pos cube.Pos, tx *world.Tx, face cube.Face) []cube.Pos {
	side := pos.Side(face)
	if side.OutOfBounds(tx.Range()) {
		return nil
	}
	positions := make([]cube.Pos, 0, 3)
	if redstoneWireDirectConnectionLoaded(tx, side, face.Opposite()) {
		positions = append(positions, side)
	}

	above := pos.Side(cube.FaceUp)
	sideAbove := side.Side(cube.FaceUp)
	if !redstoneWireBlocksConnectionLoaded(tx, above, cube.FaceDown) && redstoneWireAtLoaded(tx, sideAbove) && redstoneWireSupportedLoaded(tx, sideAbove) {
		positions = append(positions, sideAbove)
	}
	if !redstoneWireBlocksConnectionLoaded(tx, side, cube.FaceUp) {
		down := side.Side(cube.FaceDown)
		if !down.OutOfBounds(tx.Range()) && redstoneWireAtLoaded(tx, down) {
			positions = append(positions, down)
		}
	}
	return positions
}

func redstoneWireDirectConnectionLoaded(tx *world.Tx, pos cube.Pos, face cube.Face) bool {
	b, ok := tx.BlockLoaded(pos)
	if !ok {
		return false
	}
	if _, ok := b.(RedstoneWire); ok {
		return true
	}
	if _, ok := b.(world.RedstonePowerSource); ok {
		return true
	}
	if _, ok := b.(world.RedstoneStrongPowerSource); ok {
		return true
	}
	if _, ok := b.(world.RedstonePowerRelayer); ok {
		return true
	}
	return redstoneWireNonSolidComponent(pos, b, tx, face)
}

func redstoneWireNonSolidComponent(pos cube.Pos, b world.Block, tx *world.Tx, face cube.Face) bool {
	model := b.Model()
	if model == nil || model.FaceSolid(pos, face, tx) {
		return false
	}
	if _, ok := b.(world.RedstonePowerConsumer); ok {
		return true
	}
	if _, ok := b.(world.RedstonePowerTransitionConsumer); ok {
		return true
	}
	if _, ok := b.(world.RedstonePowerAction); ok {
		return true
	}
	return false
}

func redstoneWireAtLoaded(tx *world.Tx, pos cube.Pos) bool {
	b, ok := tx.BlockLoaded(pos)
	if !ok {
		return false
	}
	_, ok = b.(RedstoneWire)
	return ok
}

func redstoneWireRelevantLoaded(tx *world.Tx, pos cube.Pos) bool {
	b, ok := tx.BlockLoaded(pos)
	return ok && redstoneWireRelevant(b)
}

func redstoneWireRelevant(b world.Block) bool {
	if _, ok := b.(world.RedstonePowerSource); ok {
		return true
	}
	if _, ok := b.(world.RedstoneStrongPowerSource); ok {
		return true
	}
	if _, ok := b.(world.RedstonePowerRelayer); ok {
		return true
	}
	if _, ok := b.(world.RedstonePowerConsumer); ok {
		return true
	}
	if _, ok := b.(world.RedstonePowerTransitionConsumer); ok {
		return true
	}
	if _, ok := b.(world.RedstonePowerAction); ok {
		return true
	}
	return false
}

func redstoneWireFaceHorizontal(face cube.Face) bool {
	switch face {
	case cube.FaceNorth, cube.FaceSouth, cube.FaceWest, cube.FaceEast:
		return true
	default:
		return false
	}
}

// TrimMaterial delegates to item.RedstoneWire so the block form stays valid for smithing trim decoding too.
func (RedstoneWire) TrimMaterial() string {
	return item.RedstoneWire{}.TrimMaterial()
}

// MaterialColour delegates to item.RedstoneWire to keep trim metadata defined in one place.
func (RedstoneWire) MaterialColour() string {
	return item.RedstoneWire{}.MaterialColour()
}

// allRedstoneWires returns a list of all redstone dust states.
func allRedstoneWires() (all []world.Block) {
	for i := range 16 {
		all = append(all, RedstoneWire{Power: i})
	}
	return
}

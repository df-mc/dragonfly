package block

import (
	"time"

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

func (r RedstoneWire) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, r)
	if !used || !redstoneWireSupported(tx, pos) {
		return false
	}
	place(tx, pos, r, user, ctx)
	return placed(ctx)
}

// RedstonePower returns the wire's current signal strength from connected faces.
func (r RedstoneWire) RedstonePower(pos cube.Pos, tx *world.Tx, face cube.Face) int {
	if face == cube.FaceUp {
		return 0
	}
	if tx != nil && redstoneWireFaceHorizontal(face) && !redstoneWirePowersHorizontalFace(pos, tx, face) {
		return 0
	}
	return r.Power
}

func (RedstoneWire) RedstoneWeaklyPowersBlocks() bool {
	return true
}

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

func (r RedstoneWire) RedstonePowerUpdate(_ cube.Pos, _ *world.Tx, power int) (world.Block, bool) {
	power = world.ClampRedstonePower(power)
	if r.Power == power {
		return r, false
	}
	r.Power = power
	return r, true
}

func (r RedstoneWire) NeighbourUpdateTick(pos, _ cube.Pos, tx *world.Tx) {
	if !redstoneWireSupported(tx, pos) {
		breakBlock(r, pos, tx)
	}
}

func (RedstoneWire) HasLiquidDrops() bool {
	return true
}

func (r RedstoneWire) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(RedstoneWire{}))
}

func (RedstoneWire) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

func (r RedstoneWire) EncodeBlock() (string, map[string]any) {
	return "minecraft:redstone_wire", map[string]any{"redstone_signal": int32(world.ClampRedstonePower(r.Power))}
}

func (RedstoneWire) EncodeItem() (name string, meta int16) {
	return "minecraft:redstone", 0
}

// TrimMaterial delegates to item.RedstoneWire so the block form stays valid for smithing trim decoding too.
func (RedstoneWire) TrimMaterial() string {
	return item.RedstoneWire{}.TrimMaterial()
}

// MaterialColour delegates to item.RedstoneWire to keep trim metadata defined in one place.
func (RedstoneWire) MaterialColour() string {
	return item.RedstoneWire{}.MaterialColour()
}

func allRedstoneWires() (all []world.Block) {
	for i := range 16 {
		all = append(all, RedstoneWire{Power: i})
	}
	return
}

// redstoneTicks converts redstone ticks to a wall-clock duration at 10 redstone ticks per second.
func redstoneTicks(ticks int) time.Duration {
	return time.Duration(max(ticks, 1)) * time.Second / 10
}

// redstoneWireSupported reports whether redstone wire can stay placed at pos.
func redstoneWireSupported(tx *world.Tx, pos cube.Pos) bool {
	below := pos.Side(cube.FaceDown)
	if below.OutOfBounds(tx.Range()) {
		return false
	}
	return tx.Block(below).Model().FaceSolid(below, cube.FaceUp, tx)
}

// redstoneWireSupportedLoaded checks support without loading neighbouring chunks.
func redstoneWireSupportedLoaded(tx *world.Tx, pos cube.Pos) bool {
	below := pos.Side(cube.FaceDown)
	if below.OutOfBounds(tx.Range()) {
		return false
	}
	b, ok := tx.BlockLoaded(below)
	return ok && b.Model().FaceSolid(below, cube.FaceUp, tx)
}

// redstoneWireBlocksConnectionLoaded reports whether a loaded block blocks wire connection through face.
func redstoneWireBlocksConnectionLoaded(tx *world.Tx, pos cube.Pos, face cube.Face) bool {
	if pos.OutOfBounds(tx.Range()) {
		return true
	}
	b, ok := tx.BlockLoaded(pos)
	return ok && b.Model().FaceSolid(pos, face, tx) && redstoneConductiveBlock(pos, b, tx)
}

// redstoneWirePowersHorizontalFace reports whether wire power is exposed through a horizontal face.
func redstoneWirePowersHorizontalFace(pos cube.Pos, tx *world.Tx, face cube.Face) bool {
	return redstoneWirePoweredHorizontalFaces(pos, tx)[face]
}

// redstoneWirePoweredHorizontalFaces returns the horizontal faces connected by the wire shape.
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

// redstoneWireHorizontalConnectionPositions returns direct, step-up, and step-down connections through face.
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
		if !down.OutOfBounds(tx.Range()) && redstoneWireAtLoaded(tx, down) && redstoneWireCanTransmitDown(tx, pos) {
			positions = append(positions, down)
		}
	}
	return positions
}

// redstoneWireCanTransmitDown reports whether dust at pos may power dust one block lower.
func redstoneWireCanTransmitDown(tx *world.Tx, pos cube.Pos) bool {
	supportPos := pos.Side(cube.FaceDown)
	if supportPos.OutOfBounds(tx.Range()) {
		return false
	}
	support, ok := tx.BlockLoaded(supportPos)
	if !ok || !support.Model().FaceSolid(supportPos, cube.FaceUp, tx) {
		return false
	}
	if stepDowner, ok := support.(RedstoneWireStepDowner); ok {
		return stepDowner.CanRedstoneWireStepDown(supportPos, pos, tx)
	}
	for _, face := range cube.Faces() {
		if !support.Model().FaceSolid(supportPos, face, tx) {
			return false
		}
	}
	return true
}

// redstoneConductiveBlock reports whether b can be powered as a solid redstone conductor.
func redstoneConductiveBlock(pos cube.Pos, b world.Block, tx *world.Tx) bool {
	if _, ok := b.(world.RedstoneNonConductive); ok {
		return false
	}
	if diffuser, ok := b.(LightDiffuser); ok && diffuser.LightDiffusionLevel() == 0 {
		return false
	}
	for _, face := range cube.Faces() {
		if !b.Model().FaceSolid(pos, face, tx) {
			return false
		}
	}
	return true
}

// redstoneWireDirectConnectionLoaded reports whether a loaded block can directly connect to dust.
func redstoneWireDirectConnectionLoaded(tx *world.Tx, pos cube.Pos, face cube.Face) bool {
	b, ok := tx.BlockLoaded(pos)
	if !ok {
		return false
	}
	switch b.(type) {
	case RedstoneWire, world.RedstonePowerSource, world.RedstoneStrongPowerSource, world.RedstonePowerRelayer:
		return true
	}
	return redstoneWireNonSolidComponent(pos, b, tx, face)
}

// redstoneWireNonSolidComponent reports whether a non-solid loaded block is a redstone endpoint.
func redstoneWireNonSolidComponent(pos cube.Pos, b world.Block, tx *world.Tx, face cube.Face) bool {
	model := b.Model()
	if model == nil || model.FaceSolid(pos, face, tx) {
		return false
	}
	switch b.(type) {
	case world.RedstonePowerConsumer, world.RedstonePowerAction, world.RedstonePowerContextAction:
		return true
	}
	return false
}

// redstoneWireAtLoaded reports whether loaded block data at pos is redstone wire.
func redstoneWireAtLoaded(tx *world.Tx, pos cube.Pos) bool {
	b, ok := tx.BlockLoaded(pos)
	if !ok {
		return false
	}
	_, ok = b.(RedstoneWire)
	return ok
}

// redstoneWireRelevantLoaded reports whether a loaded block participates in redstone propagation.
func redstoneWireRelevantLoaded(tx *world.Tx, pos cube.Pos) bool {
	b, ok := tx.BlockLoaded(pos)
	return ok && redstoneWireRelevant(b)
}

// redstoneWireRelevant reports whether b participates in redstone propagation.
func redstoneWireRelevant(b world.Block) bool {
	switch b.(type) {
	case world.RedstonePowerSource,
		world.RedstoneStrongPowerSource,
		world.RedstonePowerRelayer,
		world.RedstonePowerConsumer,
		world.RedstonePowerAction,
		world.RedstonePowerContextAction:
		return true
	}
	return false
}

// redstoneWireFaceHorizontal reports whether face is one of the four horizontal faces.
func redstoneWireFaceHorizontal(face cube.Face) bool {
	switch face {
	case cube.FaceNorth, cube.FaceSouth, cube.FaceWest, cube.FaceEast:
		return true
	default:
		return false
	}
}

package block

import (
	"math/rand/v2"

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

// RedstonePower returns the wire's current signal strength from connected faces.
func (r RedstoneWire) RedstonePower(pos cube.Pos, tx *world.Tx, face cube.Face) int {
	if tx != nil && redstoneWireFaceHorizontal(face) && !redstoneWirePowersHorizontalFace(pos, tx, face) {
		return 0
	}
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

// RedstonePowerUpdate lights the lamp immediately and schedules delayed turn-off when power is removed.
func (r RedstoneLamp) RedstonePowerUpdate(pos cube.Pos, tx *world.Tx, power int) (world.Block, bool) {
	if redstoneLampPower(pos, tx, power) > 0 {
		if r.Lit {
			return r, false
		}
		r.Lit = true
		return r, true
	}
	if !r.Lit {
		return r, false
	}
	if tx != nil {
		tx.ScheduleBlockUpdate(pos, r, redstoneTicks(2))
		return r, false
	}
	r.Lit = false
	return r, true
}

// ScheduledTick turns the lamp off after its delay if it was not repowered.
func (r RedstoneLamp) ScheduledTick(pos cube.Pos, tx *world.Tx, _ *rand.Rand) {
	if tx == nil || !r.Lit || redstoneLampPower(pos, tx, tx.RedstonePower(pos)) > 0 {
		return
	}
	r.Lit = false
	tx.SetBlock(pos, r, nil)
}

func redstoneLampPower(pos cube.Pos, tx *world.Tx, power int) int {
	if tx == nil {
		return redstonePower(power)
	}
	return max(redstoneLampGraphPower(pos, tx, power), tx.RedstoneDirectPower(pos), redstoneLampConductedStrongPower(pos, tx))
}

func redstoneLampGraphPower(pos cube.Pos, tx *world.Tx, power int) int {
	power = redstonePower(power)
	if power == 0 {
		return 0
	}
	for _, face := range cube.Faces() {
		neighbour := pos.Side(face)
		if neighbour.OutOfBounds(tx.Range()) {
			continue
		}
		b, ok := tx.BlockLoaded(neighbour)
		if !ok {
			continue
		}
		if _, ok := b.(world.RedstonePowerRelayer); !ok {
			continue
		}
		if redstoneLampRelayerConnectsTo(neighbour, pos, tx, b) {
			return power
		}
	}
	return 0
}

func redstoneLampRelayerConnectsTo(relayerPos, target cube.Pos, tx *world.Tx, b world.Block) bool {
	neighbourer, ok := b.(world.RedstonePowerRelayerNeighbourer)
	if !ok {
		return true
	}
	for _, neighbour := range neighbourer.RedstoneRelayerNeighbours(relayerPos, tx) {
		if neighbour == target {
			return true
		}
	}
	return false
}

func redstoneLampConductedStrongPower(pos cube.Pos, tx *world.Tx) int {
	power := 0
	for _, face := range cube.Faces() {
		conductorPos := pos.Side(face)
		if conductorPos.OutOfBounds(tx.Range()) {
			continue
		}
		conductor, ok := tx.BlockLoaded(conductorPos)
		if !ok || !redstoneLampStrongPowerConductor(conductorPos, conductor, tx, face.Opposite()) {
			continue
		}
		power = max(power, tx.RedstoneStrongPower(conductorPos))
	}
	return redstonePower(power)
}

func redstoneLampStrongPowerConductor(pos cube.Pos, b world.Block, tx *world.Tx, face cube.Face) bool {
	if !b.Model().FaceSolid(pos, face, tx) {
		return false
	}
	if redstoneLampExplicitNonConductor(b) {
		return false
	}
	if diffuser, ok := b.(redstoneLampLightDiffuser); ok && diffuser.LightDiffusionLevel() == 0 {
		return false
	}
	return true
}

func redstoneLampExplicitNonConductor(b world.Block) bool {
	name, _ := b.EncodeBlock()
	switch name {
	case "minecraft:redstone_block", "minecraft:piston", "minecraft:sticky_piston", "minecraft:piston_arm", "minecraft:observer":
		return true
	}
	return false
}

type redstoneLampLightDiffuser interface {
	LightDiffusionLevel() uint8
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

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
	// Connections holds the way the wire connects to each of its horizontal
	// sides.
	Connections WireConnections
}

func (r RedstoneWire) UseOnBlock(pos cube.Pos, face cube.Face, _ mgl64.Vec3, tx *world.Tx, user item.User, ctx *item.UseContext) bool {
	pos, _, used := firstReplaceable(tx, pos, face, r)
	if !used || !redstoneWireSupported(tx, pos) {
		return false
	}
	r.Connections = redstoneWireConnections(pos, loadedSource{tx: tx})
	place(tx, pos, r, user, ctx)
	return placed(ctx)
}

// RedstonePower returns the wire's current signal strength from connected faces.
func (r RedstoneWire) RedstonePower(pos cube.Pos, tx *world.Tx, face cube.Face) int {
	if face == cube.FaceUp {
		return 0
	}
	if tx != nil && redstoneWireFaceHorizontal(face) && !redstoneWirePoweredHorizontalFaces(pos, loadedSource{tx: tx})[face] {
		return 0
	}
	return r.Power
}

func (RedstoneWire) RedstoneWeaklyPowersBlocks() bool {
	return true
}

func (RedstoneWire) RedstoneSignalLoss(cube.Pos, *world.Tx) int {
	return 1
}

// RedstoneRelayerNeighbours returns all wire positions directly connected to this dust, including dust stepping up or
// down adjacent blocks.
func (RedstoneWire) RedstoneRelayerNeighbours(pos cube.Pos, tx *world.Tx) []cube.Pos {
	neighbours := make([]cube.Pos, 0, 12)
	src := loadedSource{tx: tx}
	faces := redstoneWirePoweredHorizontalFaces(pos, src)
	for _, face := range cube.HorizontalFaces() {
		if !faces[face] {
			continue
		}
		side := pos.Side(face)
		if side.OutOfBounds(tx.Range()) {
			continue
		}
		if links := redstoneWireLinksThrough(pos, src, face); links.connected() {
			neighbours = append(neighbours, links.positions(side)...)
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

func (r RedstoneWire) NeighbourUpdateTick(pos, changedNeighbour cube.Pos, tx *world.Tx) {
	if !redstoneWireSupported(tx, pos) {
		breakBlock(r, pos, tx)
		return
	}
	r.updateConnections(pos, tx)
	if changedNeighbour == pos {
		// Wire one block up or down beside this wire connects to it without
		// neighbouring it, so it is not updated by the wire being placed here
		// otherwise.
		updateDiagonalRedstoneWires(pos, tx)
	}
}

// DeriveState ...
func (r RedstoneWire) DeriveState(pos cube.Pos, src world.BlockSource) world.Block {
	r.Connections = redstoneWireConnections(pos, src)
	return r
}

// updateConnections recalculates the connections of the wire at pos and
// writes the wire if they changed. Only the shape of the wire changes, which
// affects neither its neighbours nor the power it carries, so no updates
// follow the write and the power the redstone engine holds for it is kept.
func (r RedstoneWire) updateConnections(pos cube.Pos, tx *world.Tx) {
	if connections := redstoneWireConnections(pos, loadedSource{tx: tx}); connections != r.Connections {
		r.Connections = connections
		tx.SetBlock(pos, r, &world.SetOpts{DisableBlockUpdates: true, DisableRedstoneUpdates: true, KeepRedstonePower: true})
	}
}

func (RedstoneWire) HasLiquidDrops() bool {
	return true
}

func (r RedstoneWire) BreakInfo() BreakInfo {
	return newBreakInfo(0, alwaysHarvestable, nothingEffective, oneOf(RedstoneWire{})).withBreakHandler(func(pos cube.Pos, tx *world.Tx, _ item.User) {
		updateDiagonalRedstoneWires(pos, tx)
	})
}

func (RedstoneWire) SideClosed(cube.Pos, cube.Pos, *world.Tx) bool {
	return false
}

func (r RedstoneWire) EncodeBlock() (string, map[string]any) {
	properties := r.Connections.properties()
	properties["redstone_signal"] = int32(world.ClampRedstonePower(r.Power))
	return "minecraft:redstone_wire", properties
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
		for _, c := range allWireConnections() {
			all = append(all, RedstoneWire{Power: i, Connections: c})
		}
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

// loadedSource is a [world.BlockSource] reading blocks only from chunks that
// are already loaded, and air anywhere else, so that finding out what redstone
// wire connects to never loads a chunk.
type loadedSource struct{ tx *world.Tx }

// Block ...
func (s loadedSource) Block(pos cube.Pos) world.Block {
	b, _ := s.tx.BlockLoaded(pos)
	return b
}

// updateDiagonalRedstoneWires updates the connections of wire one block up or
// down beside pos, which connects to wire at pos without neighbouring it.
func updateDiagonalRedstoneWires(pos cube.Pos, tx *world.Tx) {
	src := loadedSource{tx: tx}
	for _, face := range cube.HorizontalFaces() {
		side := pos.Side(face)
		for _, diagonal := range [2]cube.Pos{side.Side(cube.FaceUp), side.Side(cube.FaceDown)} {
			if wire, ok := src.Block(diagonal).(RedstoneWire); ok {
				wire.updateConnections(diagonal, tx)
			}
		}
	}
}

// redstoneWireConnections returns the WireConnections that wire at pos is drawn
// with.
func redstoneWireConnections(pos cube.Pos, src world.BlockSource) WireConnections {
	links := redstoneWireLinksAround(pos, src)
	faces := redstoneWireShape(links, redstoneWireLinks.drawn)
	connection := func(i int) WireConnection {
		if !faces[i] {
			return NoWireConnection()
		}
		// The game only draws wire up the side of a block whose face towards
		// the wire is solid, such as stone or glass, and draws it flat towards
		// wire on top of a top slab.
		face := cube.HorizontalFaces()[i]
		if side := pos.Side(face); links[i].drawnUp && src.Block(side).Model().FaceSolid(side, face.Opposite(), src) {
			return UpWireConnection()
		}
		return SideWireConnection()
	}
	return WireConnections{north: connection(0), east: connection(1), south: connection(2), west: connection(3)}
}

// redstoneWirePoweredHorizontalFaces returns the horizontal faces connected by the wire shape.
func redstoneWirePoweredHorizontalFaces(pos cube.Pos, src world.BlockSource) map[cube.Face]bool {
	faces := redstoneWireShape(redstoneWireLinksAround(pos, src), redstoneWireLinks.connected)
	connections := make(map[cube.Face]bool, len(faces))
	for i, face := range cube.HorizontalFaces() {
		if faces[i] {
			connections[face] = true
		}
	}
	return connections
}

// redstoneWireLinksAround returns the links of the wire at pos through each
// horizontal face, in the order of [cube.HorizontalFaces].
func redstoneWireLinksAround(pos cube.Pos, src world.BlockSource) (links [4]redstoneWireLinks) {
	for i, face := range cube.HorizontalFaces() {
		links[i] = redstoneWireLinksThrough(pos, src, face)
	}
	return links
}

// redstoneWireShape returns the faces that wire with the links passed runs to,
// where linked reports if the links through a face connect the wire through it.
// Wire linked through no face runs to every face, and wire linked through one
// face also runs to the face opposite it.
func redstoneWireShape(links [4]redstoneWireLinks, linked func(redstoneWireLinks) bool) (faces [4]bool) {
	connected := 0
	for i, l := range links {
		if linked(l) {
			faces[i] = true
			connected++
		}
	}
	switch connected {
	case 0:
		faces = [4]bool{true, true, true, true}
	case 1:
		// Horizontal faces run clockwise from north, so the opposite face is
		// two further along.
		for i := range faces {
			if faces[i] {
				faces[(i+2)%4] = true
				break
			}
		}
	}
	return faces
}

// redstoneWireLinks holds what redstone wire connects to through one of its
// horizontal faces. The game draws wire with links that differ from the ones it
// carries power through: it draws wire down to wire it cannot power, such as
// from on top of a top slab or glowstone, and decides whether a block cuts off
// a climb or descent by its solidity when drawing but by whether it conducts
// power when powering, so glowstone hides a climb that still carries power.
type redstoneWireLinks struct {
	// side is true if the wire connects to the block beside it. up and down are
	// true if it powers wire on top of that block or below it, and drawnUp and
	// drawnDown if it is drawn running to that wire.
	side, up, down, drawnUp, drawnDown bool
}

// connected reports whether the wire carries power through the face.
func (l redstoneWireLinks) connected() bool {
	return l.side || l.up || l.down
}

// drawn reports whether the wire is drawn running to the face.
func (l redstoneWireLinks) drawn() bool {
	return l.side || l.drawnUp || l.drawnDown
}

// positions returns the positions linked to, where side is the position beside
// the wire.
func (l redstoneWireLinks) positions(side cube.Pos) []cube.Pos {
	positions := make([]cube.Pos, 0, 3)
	if l.side {
		positions = append(positions, side)
	}
	if l.up {
		positions = append(positions, side.Side(cube.FaceUp))
	}
	if l.down {
		positions = append(positions, side.Side(cube.FaceDown))
	}
	return positions
}

// redstoneWireLinksThrough returns the direct, step-up and step-down links of
// the wire at pos through face.
func redstoneWireLinksThrough(pos cube.Pos, src world.BlockSource, face cube.Face) redstoneWireLinks {
	side, above := pos.Side(face), pos.Side(cube.FaceUp)
	sideAbove, sideBelow := side.Side(cube.FaceUp), side.Side(cube.FaceDown)
	upper := redstoneWireAt(src, sideAbove) && redstoneWireSupportedBy(src, sideAbove)
	lower := redstoneWireAt(src, sideBelow)
	return redstoneWireLinks{
		side:      redstoneWireDirectConnection(src, side, face.Opposite()),
		up:        upper && !redstoneWireBlocksConnection(src, above, cube.FaceDown),
		down:      lower && !redstoneWireBlocksConnection(src, side, cube.FaceUp) && redstoneWireCanTransmitDown(src, pos),
		drawnUp:   upper && !redstoneWireBlocksDrawnConnection(src, above, cube.FaceDown),
		drawnDown: lower && !redstoneWireBlocksDrawnConnection(src, side, cube.FaceUp),
	}
}

// redstoneWireSupportedBy reports whether the block below pos in src can hold
// redstone wire.
func redstoneWireSupportedBy(src world.BlockSource, pos cube.Pos) bool {
	below := pos.Side(cube.FaceDown)
	return src.Block(below).Model().FaceSolid(below, cube.FaceUp, src)
}

// redstoneWireBlocksConnection reports whether the block at pos stops wire
// powering wire past it through face.
func redstoneWireBlocksConnection(src world.BlockSource, pos cube.Pos, face cube.Face) bool {
	b := src.Block(pos)
	return b.Model().FaceSolid(pos, face, src) && world.RedstoneFullPowerConductor(pos, b, src)
}

// redstoneWireBlocksDrawnConnection reports whether the block at pos stops wire
// being drawn running to wire past it through face. The game decides this by
// whether the block is solid, which differs from whether it conducts power only
// for these blocks.
func redstoneWireBlocksDrawnConnection(src world.BlockSource, pos cube.Pos, face cube.Face) bool {
	switch src.Block(pos).(type) {
	case Barrier, Glowstone, Slime:
		return true
	case Barrel, Farmland, Ice, Jukebox, TintedGlass:
		return false
	}
	return redstoneWireBlocksConnection(src, pos, face)
}

// redstoneWireCanTransmitDown reports whether dust at pos may power dust one block lower.
func redstoneWireCanTransmitDown(src world.BlockSource, pos cube.Pos) bool {
	supportPos := pos.Side(cube.FaceDown)
	support := src.Block(supportPos)
	if !support.Model().FaceSolid(supportPos, cube.FaceUp, src) {
		return false
	}
	if stepDowner, ok := support.(RedstoneWireStepDowner); ok {
		return stepDowner.CanRedstoneWireStepDown(supportPos, pos, src)
	}
	for _, face := range cube.Faces() {
		if !support.Model().FaceSolid(supportPos, face, src) {
			return false
		}
	}
	return true
}

// redstoneWireDirectConnection reports whether the block at pos can directly
// connect to dust.
func redstoneWireDirectConnection(src world.BlockSource, pos cube.Pos, face cube.Face) bool {
	b := src.Block(pos)
	switch b.(type) {
	case RedstoneWire, world.RedstonePowerSource, world.RedstoneStrongPowerSource, world.RedstonePowerRelayer:
		return true
	case Jukebox, Lectern:
		// The game connects wire to these as the redstone components they
		// are, which dragonfly does not implement for them yet.
		return true
	}
	return redstoneWireNonSolidComponent(pos, b, src, face)
}

// redstoneWireNonSolidComponent reports whether a non-solid block is a redstone
// endpoint.
func redstoneWireNonSolidComponent(pos cube.Pos, b world.Block, src world.BlockSource, face cube.Face) bool {
	model := b.Model()
	if model == nil || model.FaceSolid(pos, face, src) {
		return false
	}
	switch b.(type) {
	case world.RedstonePowerConsumer, world.RedstonePowerAction, world.RedstonePowerContextAction:
		return true
	}
	return false
}

// redstoneWireAt reports whether the block at pos is redstone wire.
func redstoneWireAt(src world.BlockSource, pos cube.Pos) bool {
	_, ok := src.Block(pos).(RedstoneWire)
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
